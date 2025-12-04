package repo

import (
	"context"
	"encoding/json"
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

	result, err := r.redis.Get(ctx, `room:`+roomName).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check room: %w", err)
	}

	if result != "" {
		return fmt.Errorf("sudah ada roomnya")
	} else {
		setRoom := r.redis.Set(context.Background(), `room:`+roomName, playerName, 300*time.Second)
		setFightRoom := r.redis.Set(context.Background(), `fightroom:`+roomName, "", 300*time.Second)
		if setRoom.Err() != nil {
			return setRoom.Err()
		}
		if setFightRoom.Err() != nil {
			return setFightRoom.Err()
		}
	}

	return nil
}
func (r *PlayRepo) JoinRoom(playerName, roomName string) error {
	ctx := context.Background()

	roomOwner, err := r.redis.Get(ctx, "room:"+roomName).Result()
	if err == redis.Nil {
		return fmt.Errorf("room not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check room: %w", err)
	}
	if roomOwner == "" {
		return fmt.Errorf("room not found")
	}

	type PlayerRoom struct {
		Player1 string `json:"player1"`
		Player2 string `json:"player2"`
	}

	var pr PlayerRoom

	jsonStr, err := r.redis.Get(ctx, "playerRoom:"+roomName).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(jsonStr), &pr)
	} else if err == redis.Nil {
		pr.Player1 = roomOwner
	} else {
		return fmt.Errorf("failed to get playerRoom: %w", err)
	}

	if pr.Player1 == playerName || pr.Player2 == playerName {
		return fmt.Errorf("player sudah ada di room")
	}

	if pr.Player2 == "" {
		pr.Player2 = playerName
	} else {
		return fmt.Errorf("room sudah penuh")
	}

	jsonBytes, _ := json.Marshal(pr)
	if err := r.redis.Set(ctx, "playerRoom:"+roomName, jsonBytes, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to save playerRoom: %w", err)
	}

	return nil
}

func (r *PlayRepo) LeaveRoom(playerName, roomName string) error {
	ctx := context.Background()

	//check user exists in room
	resultCheck, err := r.redis.Get(context.Background(), `player:`+roomName).Result()
	if err == redis.Nil {
		return fmt.Errorf("player not found in room", err)
	}
	if err != nil {
		return fmt.Errorf("failed to check player in room: %w", err)
	}
	if resultCheck != playerName {
		return fmt.Errorf("player not found")
	}
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

	if playerName == result {
		if err := r.redis.Del(ctx, `room:`+roomName).Err(); err != nil {
			return fmt.Errorf("failed to delete room: %w", err)
		}
	}

	// Delete fight room
	if err := r.redis.Del(ctx, `fightroom:`+roomName).Err(); err != nil {
		return fmt.Errorf("failed to delete fight room: %w", err)
	}

	return nil
}
