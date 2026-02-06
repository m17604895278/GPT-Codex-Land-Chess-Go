package game

import (
	"encoding/json"
	"errors"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type FlipPayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MovePayload struct {
	FromX int `json:"fromX"`
	FromY int `json:"fromY"`
	ToX   int `json:"toX"`
	ToY   int `json:"toY"`
}

func (r *Room) HandleMessage(userID string, raw []byte) error {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	switch msg.Type {
	case "flip":
		var payload FlipPayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return err
		}
		if err := r.Flip(userID, payload.X, payload.Y); err != nil {
			r.sendError(userID, err.Error())
			return err
		}
		r.broadcast("sync", r.SyncData())
	case "move":
		var payload MovePayload
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			return err
		}
		battle, err := r.Move(userID, payload.FromX, payload.FromY, payload.ToX, payload.ToY)
		if err != nil {
			r.sendError(userID, err.Error())
			return err
		}
		if battle != nil {
			r.broadcast("battle", map[string]any{
				"from":     []int{payload.FromX, payload.FromY},
				"to":       []int{payload.ToX, payload.ToY},
				"attacker": battle.AttackerType,
				"defender": battle.DefenderType,
				"result":   battle.Result,
			})
		}
		if r.Status == StatusFinished {
			r.broadcast("game_over", map[string]any{
				"winner": r.Winner,
				"reason": r.Reason,
			})
			return nil
		}
		r.broadcast("sync", r.SyncData())
	case "ping":
		r.sendTo(userID, "pong", map[string]any{"ts": nowUnix()})
	default:
		r.sendError(userID, "unknown message type")
		return errors.New("unknown message type")
	}
	return nil
}

func (r *Room) broadcast(msgType string, data map[string]any) {
	if r.Player1 != nil {
		_ = r.Player1.Conn.WriteJSON(map[string]any{
			"type": msgType,
			"data": data,
		})
	}
	if r.Player2 != nil {
		_ = r.Player2.Conn.WriteJSON(map[string]any{
			"type": msgType,
			"data": data,
		})
	}
}

func (r *Room) sendTo(userID, msgType string, data map[string]any) {
	player, err := r.playerByID(userID)
	if err != nil || player.Conn == nil {
		return
	}
	_ = player.Conn.WriteJSON(map[string]any{
		"type": msgType,
		"data": data,
	})
}

func (r *Room) sendError(userID, msg string) {
	r.sendTo(userID, "error", map[string]any{"msg": msg})
}
