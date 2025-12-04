package main

import (
	configs "guntingbatukertas/config/redis"
	"guntingbatukertas/controller/websocket"
	"guntingbatukertas/repo"
	"guntingbatukertas/service"
	"net/http"
)

func main() {

	redisClient := configs.RedisConfig()

	playRepo := repo.NewPlayRepo(redisClient)
	playService := service.NewPlayService(*&playRepo)
	websocket := websocket.NewWebSocketHandler(playService)
	http.HandleFunc("/ws", websocket.CreateRoom)
	http.HandleFunc("/join", websocket.JoinRoom)
	http.HandleFunc("/leave", websocket.LeaveRoom)
	http.HandleFunc("/fight", websocket.FightRoom)
	http.ListenAndServe(":8082", nil)

}
