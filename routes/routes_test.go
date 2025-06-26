package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"go-fiber-api/models"
	"go-fiber-api/repositories"
	"go-fiber-api/routes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/inmemory"
)

func setupTestApp(t *testing.T) (*fiber.App, context.Context) {
	ctx := context.TODO()

	// Tạo MongoDB in-memory
	memSrv, err := inmemory.New(ctx)
	assert.NoError(t, err)

	db := memSrv.Database("testdb")

	app := fiber.New()
	routes.Setup(app, db)

	return app, ctx
}

func TestIntegration_CreateUser_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	payload := map[string]interface{}{
		"username":    "testuser",
		"password":    "123456",
		"role":        "admin",
		"name":        "Tester",
		"role_groups": []string{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_CreateUser_Duplicate(t *testing.T) {
	app, ctx := setupTestApp(t)

	// Insert trước bản ghi
	userRepo := repositories.NewUserRepository(app.Config().Config.Context.(*inmemory.Database))
	_ = userRepo.Create(ctx, &models.User{
		Username:   "testuser",
		Password:   "hashedpassword",
		Role:       "admin",
		Name:       "Tester",
		RoleGroups: []primitive.ObjectID{},
	})

	payload := map[string]interface{}{
		"username":    "testuser",
		"password":    "123456",
		"role":        "admin",
		"name":        "Tester",
		"role_groups": []string{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_GetUsersByRole(t *testing.T) {
	app, ctx := setupTestApp(t)

	// Insert user
	userRepo := repositories.NewUserRepository(app.Config().Config.Context.(*inmemory.Database))
	_ = userRepo.Create(ctx, &models.User{
		Username:   "member1",
		Password:   "hashed",
		Role:       "member",
		Name:       "Member One",
		RoleGroups: []primitive.ObjectID{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users?role=member", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
