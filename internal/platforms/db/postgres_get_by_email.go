package db

import (
	"context"

	"go-guardiao-api/pkg/models"
)

// GetUserByEmail retorna o usuário pelo email.
// Retorna pgx.ErrNoRows se não existir.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	u := models.User{}
	sql := `SELECT id, email, name, theme FROM users WHERE email = $1`
	err := c.pool.QueryRow(ctx, sql, email).Scan(&u.ID, &u.Email, &u.Name, &u.Theme)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}
