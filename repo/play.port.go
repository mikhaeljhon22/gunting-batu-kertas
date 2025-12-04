package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PlayRepo struct {
	redis *redis.Client
}

func NewPlayRepo(redis *redis.Client) *PlayRepo {
	return &PlayRepo{redis: redis}
}

func (r *PlayRepo) CreateRoom(playerName, roomName string) error {
	ctx := context.Background()
	
	// Check if room already exists
	result, err := r.redis.Get(ctx, `room:`+roomName).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check room: %w", err)
	}
	
	if result != "" {
		return fmt.Errorf("room already exists")
	}
	
	// Create room
	if err := r.redis.Set(ctx, `room:`+roomName, playerName, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	
	// Initialize fight room
	if err := r.redis.Set(ctx, `fightroom:`+roomName, "", 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to create fight room: %w", err)
	}
	
	return nil
}

func (r *PlayRepo) JoinRoom(playerName, roomName string) error {
	ctx := context.Background()
	
	// Check if room exists
	result, err := r.redis.Get(ctx, `room:`+roomName).Result()
	if err == redis.Nil {
		return fmt.Errorf("room not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check room: %w", err)
	}
	
	if result == "" {
		return fmt.Errorf("room not found")
	}
	
	// Add player to room
	if err := r.redis.Set(ctx, `player:`+roomName, playerName, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}
	
	return nil
}

func (r *PlayRepo) LeaveRoom(playerName, roomName string) error {
	ctx := context.Background()
	
	// Check if room exists
	result, err := r.redis.Get(ctx, `room:`+roomName).Result()
	if err == redis.Nil {
		return fmt.Errorf("room not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check room: %w", err)
	}
	
	if result == "" {
		return fmt.Errorf("room not found")
	}
	
	// Delete player
	if err := r.redis.Del(ctx, `player:`+roomName).Err(); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}
	
	// Delete room
	if err := r.redis.Del(ctx, `room:`+roomName).Err(); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	
	// Delete fight room
	if err := r.redis.Del(ctx, `fightroom:`+roomName).Err(); err != nil {
		return fmt.Errorf("failed to delete fight room: %w", err)
	}
	
	return nil
}