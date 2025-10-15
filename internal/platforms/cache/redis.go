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

var ErrCacheMiss = errors.New("cache miss")

type Client struct {
	rdb *redis.Client
}

func NewCacheClient(addr, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("AVISO: Falha ao conectar ao Redis (Cache está desativado): %v", err)
		return nil, err
	}

	log.Println("Conexão com Redis estabelecida com sucesso.")
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() {
	if c != nil && c.rdb != nil {
		if err := c.rdb.Close(); err != nil {
			log.Printf("ERRO: Falha ao fechar conexão com Redis: %v", err)
		} else {
			log.Println("Conexão com Redis fechada.")
		}
	}
}

func (c *Client) GetManaBalance(ctx context.Context, userID string) (int, error) {
	if c == nil || c.rdb == nil {
		return 0, errors.New("redis client não está conectado")
	}
	key := fmt.Sprintf("mana:%s", userID)
	val, err := c.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, ErrCacheMiss
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Client) SetManaBalance(ctx context.Context, userID string, balance int) error {
	if c == nil || c.rdb == nil {
		return errors.New("redis client não está conectado")
	}
	key := fmt.Sprintf("mana:%s", userID)
	return c.rdb.Set(ctx, key, balance, 1*time.Hour).Err()
}

func (c *Client) UpdateLeaderboard(ctx context.Context, userID string, mana int) error {
	if c == nil || c.rdb == nil {
		return errors.New("redis client não está conectado")
	}
	key := "global_leaderboard"
	return c.rdb.ZAdd(ctx, key, redis.Z{Score: float64(mana), Member: userID}).Err()
}

func (c *Client) GetLeaderboard(ctx context.Context, limit int64) ([]models.LeaderboardEntry, error) {
	if c == nil || c.rdb == nil {
		return nil, errors.New("redis client não está conectado")
	}
	key := "global_leaderboard"
	results, err := c.rdb.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]models.LeaderboardEntry, 0, len(results))
	for _, z := range results {
		userID, ok := z.Member.(string)
		if !ok {
			continue // ignora entradas inválidas
		}
		entries = append(entries, models.LeaderboardEntry{
			UserID:   userID,
			UserName: fmt.Sprintf("User-%s", userID), // Mock do nome
			Mana:     int(z.Score),
		})
	}
	return entries, nil
}
