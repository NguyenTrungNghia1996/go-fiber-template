package seed

import (
    "context"
    "log"
    "time"

    "go-fiber-api/models"
    "go-fiber-api/repositories"

    "golang.org/x/crypto/bcrypt"
)

// SeedSuperAdmin ensures a default super admin account exists.
func SeedSuperAdmin(repo *repositories.SuperAdminRepository) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    exist, err := repo.FindByUsername(ctx, "sa")
    if err != nil {
        return err
    }
    if exist != nil {
        log.Println("Seed: default super admin 'sa' already exists")
        return nil
    }
    hash, err := bcrypt.GenerateFromPassword([]byte("sa123"), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    sa := &models.SuperAdmin{
        Username:     "sa",
        PasswordHash: string(hash),
        Name:         "Super Admin",
        Email:        "sa@example.com",
    }
    if err := repo.Create(ctx, sa); err != nil {
        return err
    }
    log.Println("Seed: created default super admin 'sa'")
    return nil
}

