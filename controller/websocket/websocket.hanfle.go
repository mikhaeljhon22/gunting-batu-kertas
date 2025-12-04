package websocket

import (
	"encoding/json"
	"fmt"
	"guntingbatukertas/service"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	playService *service.PlayService
}

func NewWebSocketHandler(ps *service.PlayService) *WebSocketHandler {
	return &WebSocketHandler{
		playService: ps,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WebSocketHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	type CreateRoomPayload struct {
		PlayerName string `json:"playerName"`
		RoomName   string `json:"roomName"`
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		var payload CreateRoomPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			fmt.Println("Invalid JSON:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`))
			continue
		}

		err = h.playService.CreateRoom(payload.PlayerName, payload.RoomName)
		if err != nil {
			response := map[string]string{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, jsonResponse)
			continue
		}

		response := map[string]string{"message": "room created successfully", "roomName": payload.RoomName, "playerName": payload.PlayerName}
		jsonResponse, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, jsonResponse)
	}
}

func (h *WebSocketHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	type JoinRoomPayload struct {
		PlayerName string `json:"playerName"`
		RoomName   string `json:"roomName"`
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		var payload JoinRoomPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			fmt.Println("Invalid JSON:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`))
			continue
		}

		err = h.playService.JoinRoom(payload.PlayerName, payload.RoomName)
		if err != nil {
			response := map[string]string{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, jsonResponse)
			continue
		}

		response := map[string]string{"message": "joined room successfully", "roomName": payload.RoomName, "playerName": payload.PlayerName}
		jsonResponse, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, jsonResponse)
	}
}

func (h *WebSocketHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	type LeaveRoomPayload struct {
		PlayerName string `json:"playerName"`
		RoomName   string `json:"roomName"`
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		var payload LeaveRoomPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			fmt.Println("Invalid JSON:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`))
			continue
		}

		err = h.playService.LeaveRoom(payload.PlayerName, payload.RoomName)
		if err != nil {
			response := map[string]string{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, jsonResponse)
			continue
		}

		response := map[string]string{"message": "left room successfully", "roomName": payload.RoomName, "playerName": payload.PlayerName}
		jsonResponse, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, jsonResponse)
	}
}

func (h *WebSocketHandler) FightRoom(w http.ResponseWriter, r *http.Request) {
	type FightPayload struct {
		PlayerName string `json:"playerName"`
		Move       string `json:"move"`
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		var payload FightPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			fmt.Println("Invalid JSON:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`))
			continue
		}

		err = h.playService.SetPlayerMove(payload.PlayerName, payload.Move)
		if err != nil {
			response := map[string]string{"error": err.Error()}
			jsonResponse, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, jsonResponse)
			continue
		}

		result, err := h.playService.GetFightResult(payload.PlayerName)
		if err != nil {
			response := map[string]string{
				"status":  "waiting",
				"message": err.Error(),
			}
			jsonResponse, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, jsonResponse)
			continue
		}

		response := map[string]interface{}{
			"status": "complete",
			"result": result,
		}
		jsonResponse, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, jsonResponse)
	}
}
