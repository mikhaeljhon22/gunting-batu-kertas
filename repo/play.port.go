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

type PlayerRoom struct {
	Player1 string `json:"player1"`
	Player2 string `json:"player2"`
}

func (r *PlayRepo) CreateRoom(playerName, roomName string) error {
	ctx := context.Background()

	result, err := r.redis.Get(ctx, "room:"+roomName).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check room: %w", err)
	}

	if result != "" {
		return fmt.Errorf("sudah ada roomnya")
	}

	if err := r.redis.Set(ctx, "room:"+roomName, playerName, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	pr := PlayerRoom{
		Player1: playerName,
		Player2: "",
	}
	jsonBytes, _ := json.Marshal(pr)
	if err := r.redis.Set(ctx, "playerRoom:"+roomName, jsonBytes, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to create playerRoom: %w", err)
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

	result, err := r.redis.Get(ctx, "room:"+roomName).Result()
	if err == redis.Nil {
		return fmt.Errorf("room not found")
	}
	if err != nil {
		return fmt.Errorf("failed to check room: %w", err)
	}
	if result == "" {
		return fmt.Errorf("room not found")
	}

	var pr PlayerRoom
	jsonStr, err := r.redis.Get(ctx, "playerRoom:"+roomName).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(jsonStr), &pr)
	}

	if pr.Player1 != playerName && pr.Player2 != playerName {
		return fmt.Errorf("player not found in room")
	}

	if playerName == result {
		if err := r.redis.Del(ctx, "room:"+roomName).Err(); err != nil {
			return fmt.Errorf("failed to delete room: %w", err)
		}
		if err := r.redis.Del(ctx, "playerRoom:"+roomName).Err(); err != nil {
			return fmt.Errorf("failed to delete playerRoom: %w", err)
		}
	} else {
		pr.Player2 = ""
		jsonBytes, _ := json.Marshal(pr)
		if err := r.redis.Set(ctx, "playerRoom:"+roomName, jsonBytes, 300*time.Second).Err(); err != nil {
			return fmt.Errorf("failed to update playerRoom: %w", err)
		}
	}

	return nil
}

func (r *PlayRepo) GetPlayerRoom(playerName string) (string, error) {
	ctx := context.Background()

	keys, err := r.redis.Keys(ctx, "playerRoom:*").Result()
	if err != nil {
		return "", fmt.Errorf("failed to search rooms: %w", err)
	}

	for _, key := range keys {
		jsonStr, err := r.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var pr PlayerRoom
		if err := json.Unmarshal([]byte(jsonStr), &pr); err != nil {
			continue
		}

		if pr.Player1 == playerName || pr.Player2 == playerName {
			roomName := key[11:]
			return roomName, nil
		}
	}

	return "", fmt.Errorf("player not in any room")
}

func (r *PlayRepo) SetPlayerMove(playerName, move string) error {
	ctx := context.Background()

	roomName, err := r.GetPlayerRoom(playerName)
	if err != nil {
		return err
	}

	if move != "gunting" && move != "batu" && move != "kertas" {
		return fmt.Errorf("invalid move, must be: gunting, batu, or kertas")
	}

	key := fmt.Sprintf("move:%s:%s", roomName, playerName)
	if err := r.redis.Set(ctx, key, move, 300*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to set move: %w", err)
	}

	return nil
}

func (r *PlayRepo) GetFightResult(playerName string) (map[string]interface{}, error) {
	ctx := context.Background()

	roomName, err := r.GetPlayerRoom(playerName)
	if err != nil {
		return nil, err
	}

	var pr PlayerRoom
	jsonStr, err := r.redis.Get(ctx, "playerRoom:"+roomName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get playerRoom: %w", err)
	}
	if err := json.Unmarshal([]byte(jsonStr), &pr); err != nil {
		return nil, fmt.Errorf("failed to parse playerRoom: %w", err)
	}

	if pr.Player2 == "" {
		return nil, fmt.Errorf("waiting for player 2 to join")
	}

	move1, err := r.redis.Get(ctx, fmt.Sprintf("move:%s:%s", roomName, pr.Player1)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("waiting for %s to make a move", pr.Player1)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get move 1: %w", err)
	}

	move2, err := r.redis.Get(ctx, fmt.Sprintf("move:%s:%s", roomName, pr.Player2)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("waiting for %s to make a move", pr.Player2)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get move 2: %w", err)
	}

	winner := determineWinner(move1, move2, pr.Player1, pr.Player2)

	if winner != "draw" {
		r.addPoint(ctx, winner)
	}

	point1, _ := r.GetPlayerPoint(pr.Player1)
	point2, _ := r.GetPlayerPoint(pr.Player2)

	r.redis.Del(ctx, fmt.Sprintf("move:%s:%s", roomName, pr.Player1))
	r.redis.Del(ctx, fmt.Sprintf("move:%s:%s", roomName, pr.Player2))

	return map[string]interface{}{
		"roomName": roomName,
		"player1":  pr.Player1,
		"player2":  pr.Player2,
		"move1":    move1,
		"move2":    move2,
		"winner":   winner,
		"points": map[string]int{
			pr.Player1: point1,
			pr.Player2: point2,
		},
	}, nil
}

func (r *PlayRepo) addPoint(ctx context.Context, playerName string) error {
	key := playerName + "_poin"
	return r.redis.Incr(ctx, key).Err()
}

func (r *PlayRepo) GetPlayerPoint(playerName string) (int, error) {
	ctx := context.Background()
	key := playerName + "_poin"

	result, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var point int
	fmt.Sscanf(result, "%d", &point)
	return point, nil
}

func determineWinner(move1, move2, player1, player2 string) string {
	if move1 == move2 {
		return "draw"
	}

	if (move1 == "gunting" && move2 == "kertas") ||
		(move1 == "batu" && move2 == "gunting") ||
		(move1 == "kertas" && move2 == "batu") {
		return player1
	}

	return player2
}
