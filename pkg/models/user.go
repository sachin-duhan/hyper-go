package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (*User, error) {
	var user User
	err := pool.QueryRow(ctx,
		"SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uint) (*User, error) {
	var user User
	err := pool.QueryRow(ctx,
		"SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1",
		id).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		user.Email, string(hashedPassword), user.Role, now, now).Scan(&user.ID)
	if err != nil {
		return err
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	user.Password = "" // Clear password before returning
	return nil
}

func GetAllUsers(ctx context.Context, pool *pgxpool.Pool) ([]User, error) {
	rows, err := pool.Query(ctx,
		"SELECT id, email, role, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func ValidateUserCredentials(ctx context.Context, pool *pgxpool.Pool, email, password string) (*User, error) {
	user, err := GetUserByEmail(ctx, pool, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	user.Password = "" // Clear password before returning
	return user, nil
}
