package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type defaultUserSeed struct {
	Name        string
	Email       string
	PhoneNumber string
	Role        string
	Password    string
}

var defaultUsers = []defaultUserSeed{
	{
		Name:        "Admin",
		Email:       "admin@gmail.com",
		PhoneNumber: "+628000000001",
		Role:        "admin",
		Password:    "Ticketing@123",
	},
	{
		Name:        "Agent",
		Email:       "agent@gamil.com",
		PhoneNumber: "+628000000002",
		Role:        "agent",
		Password:    "Ticketing@123",
	},
	{
		Name:        "Requester",
		Email:       "requester@gamil.com",
		PhoneNumber: "+628000000003",
		Role:        "requester",
		Password:    "Ticketing@123",
	},
}

const querySeedUser = `
	INSERT INTO users (name, email, phonenumber, role, password_hash)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT DO NOTHING
`

// SeedDefaultUsers inserts initial admin and agent accounts if they do not already exist.
func SeedDefaultUsers(ctx context.Context, pool *pgxpool.Pool) error {
	for _, u := range defaultUsers {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash seed password for %s: %w", u.Email, err)
		}

		if _, err := pool.Exec(ctx, querySeedUser, u.Name, u.Email, u.PhoneNumber, u.Role, string(hash)); err != nil {
			return fmt.Errorf("seed user %s: %w", u.Email, err)
		}
	}

	return nil
}
