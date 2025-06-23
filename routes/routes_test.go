package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"go-fiber-api/controllers"
	"go-fiber-api/middleware"
	"go-fiber-api/models"
	"go-fiber-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memoryUserRepo struct {
	mu    sync.Mutex
	users map[string]*models.User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{users: make(map[string]*models.User)}
}

func (m *memoryUserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username {
			c := *u
			return &c, nil
		}
	}
	return nil, fiber.ErrNotFound
}

func (m *memoryUserRepo) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ID = primitive.NewObjectID()
	cp := *user
	m.users[user.ID.Hex()] = &cp
	return nil
}

func (m *memoryUserRepo) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryUserRepo) GetByRole(ctx context.Context, role string) ([]models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []models.User
	for _, u := range m.users {
		if role == "" || u.Role == role {
			cp := *u
			cp.Password = ""
			res = append(res, cp)
		}
	}
	return res, nil
}

func (m *memoryUserRepo) UpdatePassword(ctx context.Context, id string, hashed string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		u.Password = hashed
		return nil
	}
	return fiber.ErrNotFound
}

func (m *memoryUserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		c := *u
		return &c, nil
	}
	return nil, fiber.ErrNotFound
}

func setupApp(t *testing.T) (*fiber.App, *memoryUserRepo) {
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("MINIO_ENDPOINT", "localhost:9000")
	os.Setenv("MINIO_ACCESS_KEY", "key")
	os.Setenv("MINIO_SECRET_KEY", "secret")
	os.Setenv("MINIO_BUCKET", "bucket")
	os.Setenv("MINIO_SSL", "false")

	repo := newMemoryUserRepo()
	controllers.SetLoginRepo(repo)
	userCtrl := controllers.NewUserController(repo)

	app := fiber.New()
	app.Post("/login", controllers.Login)
	app.Get("/test", controllers.Hello)

	api := app.Group("/api", middleware.Protected())
	api.Get("/test2", controllers.Hello)
	api.Get("/me", userCtrl.GetCurrentUser)
	api.Put("/users/password", userCtrl.ChangeUserPassword)
	api.Put("/presigned_url", controllers.GetUploadUrl)

	admin := api.Group("/users", middleware.AdminOnly())
	admin.Post("/", userCtrl.CreateUser)
	admin.Get("/", userCtrl.GetUsersByRole)

	return app, repo
}

func TestRoutes(t *testing.T) {
	app, repo := setupApp(t)

	// seed admin user
	hashed, _ := utils.HashPassword("admin123")
	repo.Create(context.TODO(), &models.User{Username: "admin", Password: hashed, Role: "admin"})

	// Hello public
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// login
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
			ID    string `json:"id"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Data.Token

	// access protected route
	req = httptest.NewRequest(http.MethodGet, "/api/test2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// create user
	payload := map[string]string{"username": "user1", "password": "pass", "role": "member"}
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// duplicate
	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// list by role
	req = httptest.NewRequest(http.MethodGet, "/api/users?role=member", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// get current user
	req = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// change password
	body, _ = json.Marshal(map[string]string{"old_password": "admin123", "new_password": "newpass"})
	req = httptest.NewRequest(http.MethodPut, "/api/users/password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// presigned url
	body, _ = json.Marshal(models.PutObjectUpload{Key: "test.txt"})
	req = httptest.NewRequest(http.MethodPut, "/api/presigned_url", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
