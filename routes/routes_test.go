package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go-fiber-api/controllers"
	"go-fiber-api/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mockUserRepo struct {
	users map[string]models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]models.User)}
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	if u, ok := m.users[username]; ok {
		return &u, nil
	}
	return nil, mongo.ErrNoDocuments
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.Username]; exists {
		return errors.New("exists")
	}
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	m.users[user.Username] = *user
	return nil
}

func (m *mockUserRepo) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	_, ok := m.users[username]
	return ok, nil
}

func (m *mockUserRepo) GetByRole(ctx context.Context, role string) ([]models.User, error) {
	var res []models.User
	for _, u := range m.users {
		if role == "" || u.Role == role {
			u.Password = ""
			res = append(res, u)
		}
	}
	return res, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id string, hashed string) error {
	for k, u := range m.users {
		if u.ID.Hex() == id {
			u.Password = hashed
			m.users[k] = u
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID.Hex() == id {
			return &u, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func setupApp() (*fiber.App, *mockUserRepo) {
	repo := newMockUserRepo()
	ctrl := controllers.NewUserController(repo)
	app := fiber.New()
	app.Post("/api/users", ctrl.CreateUser)
	app.Get("/api/users", ctrl.GetUsersByRole)
	return app, repo
}

func TestCreateUserSuccess(t *testing.T) {
	app, _ := setupApp()
	payload := models.User{Username: "test", Password: "123", Role: "admin"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateUserDuplicate(t *testing.T) {
	app, repo := setupApp()
	repo.users["test"] = models.User{ID: primitive.NewObjectID(), Username: "test", Password: "hash", Role: "admin"}
	payload := models.User{Username: "test", Password: "123", Role: "admin"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetUsersByRole(t *testing.T) {
	app, repo := setupApp()
	repo.users["member"] = models.User{ID: primitive.NewObjectID(), Username: "member", Password: "hash", Role: "member"}
	req := httptest.NewRequest(http.MethodGet, "/api/users?role=member", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
