package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-guardiao-api/pkg/models"

	"github.com/redis/go-redis/v9"
)

// Client é a estrutura que contém a conexão real com o Redis.
type Client struct {
	rdb *redis.Client
}

// NewCacheClient estabelece a conexão com o Redis.
func NewCacheClient(addr string, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica a conexão
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("AVISO: Falha ao conectar ao Redis (Cache está desativado): %v", err)
		// Em desenvolvimento, permitimos que a app continue rodando, mas sem cache.
		return &Client{rdb: nil}, nil
	}

	log.Println("Conexão com Redis estabelecida com sucesso.")
	return &Client{rdb: rdb}, nil
}

// Close fecha a conexão com o Redis.
func (c *Client) Close() {
	if c.rdb != nil {
		if err := c.rdb.Close(); err != nil {
			log.Printf("ERRO: Falha ao fechar conexão com Redis: %v", err)
			return
		}
		log.Println("Conexão com Redis fechada.")
	}
}

// GetManaBalance consulta o saldo de Mana no cache.
func (c *Client) GetManaBalance(ctx context.Context, userID string) (int, error) {
	if c.rdb == nil {
		return 0, fmt.Errorf("redis client não está conectado")
	}

	key := fmt.Sprintf("mana:%s", userID)
	val, err := c.rdb.Get(ctx, key).Int()

	if errors.Is(err, redis.Nil) {
		return 0, fmt.Errorf("cache miss") // Cache miss: buscar no DB
	}
	if err != nil {
		log.Printf("ERRO Cache: Falha ao buscar Mana no Redis: %v", err)
		return 0, err
	}
	return val, nil
}

// SetManaBalance atualiza o saldo de Mana no cache.
func (c *Client) SetManaBalance(ctx context.Context, userID string, balance int) error {
	if c.rdb == nil {
		return fmt.Errorf("redis client não está conectado")
	}

	key := fmt.Sprintf("mana:%s", userID)
	// Expiração de 1 hora
	return c.rdb.Set(ctx, key, balance, 1*time.Hour).Err()
}

// UpdateLeaderboard atualiza o placar de líderes (usando Sorted Sets).
func (c *Client) UpdateLeaderboard(ctx context.Context, userID string, mana int) error {
	if c.rdb == nil {
		return fmt.Errorf("redis client não está conectado")
	}
	key := "global_leaderboard"
	return c.rdb.ZAdd(ctx, key, redis.Z{Score: float64(mana), Member: userID}).Err()
}

// GetLeaderboard busca o top N do placar.
func (c *Client) GetLeaderboard(ctx context.Context, limit int64) ([]models.LeaderboardEntry, error) {
	if c.rdb == nil {
		return nil, fmt.Errorf("redis client não está conectado")
	}

	key := "global_leaderboard"
	results, err := c.rdb.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]models.LeaderboardEntry, len(results))
	for i, z := range results {
		entries[i] = models.LeaderboardEntry{
			UserID:   z.Member.(string),
			UserName: fmt.Sprintf("User-%s", z.Member.(string)), // Mock do nome
			Mana:     int(z.Score),
		}
	}
	return entries, nil
}
