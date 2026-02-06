package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	BoardRows = 12
	BoardCols = 5
)

const (
	CampUnknown = "unknown"
	CampRed     = "red"
	CampBlue    = "blue"
)

const (
	StatusWaiting  = "waiting"
	StatusPlaying  = "playing"
	StatusFinished = "finished"
)

const (
	PieceFlag     = "军旗"
	PieceMine     = "地雷"
	PieceBomb     = "炸弹"
	PieceEngineer = "工兵"
)

type WebSocketConn interface {
	WriteJSON(v any) error
}

type Piece struct {
	ID      string
	Type    string
	Camp    string
	Rank    int
	X       int
	Y       int
	Flipped bool
	Alive   bool
}

type Player struct {
	UserID string
	Camp   string
	Online bool
	Conn   WebSocketConn
}

type Cell struct {
	PieceID  string
	Walkable bool
}

type Board struct {
	Rows  int
	Cols  int
	Cells [][]Cell
}

func NewBoard(rows, cols int) *Board {
	cells := make([][]Cell, rows)
	for i := 0; i < rows; i++ {
		cells[i] = make([]Cell, cols)
		for j := 0; j < cols; j++ {
			cells[i][j] = Cell{Walkable: true}
		}
	}
	return &Board{Rows: rows, Cols: cols, Cells: cells}
}

func (b *Board) InBounds(x, y int) bool {
	return x >= 0 && x < b.Cols && y >= 0 && y < b.Rows
}

func (b *Board) GetCell(x, y int) (Cell, error) {
	if !b.InBounds(x, y) {
		return Cell{}, fmt.Errorf("out of bounds")
	}
	return b.Cells[y][x], nil
}

func (b *Board) SetPiece(x, y int, pieceID string) error {
	if !b.InBounds(x, y) {
		return fmt.Errorf("out of bounds")
	}
	b.Cells[y][x].PieceID = pieceID
	return nil
}

type Room struct {
	RoomID  string
	Player1 *Player
	Player2 *Player
	Board   *Board
	Pieces  map[string]*Piece
	Turn    string
	Status  string
	Winner  string
	Reason  string
	Step    int
}

func NewRoom(roomID string, player1, player2 *Player, pieces map[string]*Piece) *Room {
	board := NewBoard(BoardRows, BoardCols)
	return &Room{
		RoomID:  roomID,
		Player1: player1,
		Player2: player2,
		Board:   board,
		Pieces:  pieces,
		Turn:    CampUnknown,
		Status:  StatusWaiting,
		Step:    0,
	}
}

func (r *Room) Start(turn string) {
	r.Turn = turn
	r.Status = StatusPlaying
}

func (r *Room) currentPlayer() (*Player, error) {
	if r.Turn == r.Player1.Camp {
		return r.Player1, nil
	}
	if r.Turn == r.Player2.Camp {
		return r.Player2, nil
	}
	return nil, fmt.Errorf("turn camp not set")
}

func (r *Room) playerByID(userID string) (*Player, error) {
	if r.Player1 != nil && r.Player1.UserID == userID {
		return r.Player1, nil
	}
	if r.Player2 != nil && r.Player2.UserID == userID {
		return r.Player2, nil
	}
	return nil, fmt.Errorf("player not found")
}

func (r *Room) opponentCamp(camp string) string {
	if camp == CampRed {
		return CampBlue
	}
	if camp == CampBlue {
		return CampRed
	}
	return CampUnknown
}

func (r *Room) Flip(userID string, x, y int) error {
	if r.Status != StatusPlaying {
		return errors.New("room not playing")
	}
	player, err := r.playerByID(userID)
	if err != nil {
		return err
	}
	if player.Camp != CampUnknown && player.Camp != r.Turn {
		return errors.New("not your turn")
	}
	cell, err := r.Board.GetCell(x, y)
	if err != nil {
		return err
	}
	if cell.PieceID == "" {
		return errors.New("no piece to flip")
	}
	piece := r.Pieces[cell.PieceID]
	if piece == nil {
		return errors.New("piece not found")
	}
	if piece.Flipped {
		return errors.New("piece already flipped")
	}
	piece.Flipped = true
	if player.Camp == CampUnknown {
		player.Camp = piece.Camp
		opponent := r.opponentCamp(piece.Camp)
		if r.Player1 == player {
			r.Player2.Camp = opponent
		} else {
			r.Player1.Camp = opponent
		}
		if r.Turn == CampUnknown {
			r.Turn = piece.Camp
		}
	}
	r.advanceTurn()
	return nil
}

func (r *Room) Move(userID string, fromX, fromY, toX, toY int) (*BattleResult, error) {
	if r.Status != StatusPlaying {
		return nil, errors.New("room not playing")
	}
	player, err := r.playerByID(userID)
	if err != nil {
		return nil, err
	}
	if player.Camp != r.Turn {
		return nil, errors.New("not your turn")
	}
	if !r.Board.InBounds(fromX, fromY) || !r.Board.InBounds(toX, toY) {
		return nil, errors.New("out of bounds")
	}
	if abs(fromX-toX)+abs(fromY-toY) != 1 {
		return nil, errors.New("invalid move")
	}
	fromCell, _ := r.Board.GetCell(fromX, fromY)
	if fromCell.PieceID == "" {
		return nil, errors.New("no piece to move")
	}
	piece := r.Pieces[fromCell.PieceID]
	if piece == nil || !piece.Alive {
		return nil, errors.New("piece not available")
	}
	if !piece.Flipped {
		return nil, errors.New("piece not flipped")
	}
	if piece.Camp != player.Camp {
		return nil, errors.New("cannot move opponent piece")
	}
	if piece.Type == PieceFlag || piece.Type == PieceMine {
		return nil, errors.New("piece cannot move")
	}
	toCell, _ := r.Board.GetCell(toX, toY)
	if toCell.PieceID == "" {
		if err := r.Board.SetPiece(fromX, fromY, ""); err != nil {
			return nil, err
		}
		if err := r.Board.SetPiece(toX, toY, piece.ID); err != nil {
			return nil, err
		}
		piece.X = toX
		piece.Y = toY
		r.advanceTurn()
		return nil, nil
	}
	defender := r.Pieces[toCell.PieceID]
	if defender == nil || !defender.Alive {
		return nil, errors.New("defender not available")
	}
	if defender.Camp == player.Camp {
		return nil, errors.New("cannot attack own piece")
	}
	result := ResolveBattle(piece, defender)
	switch {
	case result.AttackerAlive && !result.DefenderAlive:
		if err := r.Board.SetPiece(fromX, fromY, ""); err != nil {
			return nil, err
		}
		if err := r.Board.SetPiece(toX, toY, piece.ID); err != nil {
			return nil, err
		}
		piece.X = toX
		piece.Y = toY
	case !result.AttackerAlive && result.DefenderAlive:
		if err := r.Board.SetPiece(fromX, fromY, ""); err != nil {
			return nil, err
		}
	case !result.AttackerAlive && !result.DefenderAlive:
		if err := r.Board.SetPiece(fromX, fromY, ""); err != nil {
			return nil, err
		}
		if err := r.Board.SetPiece(toX, toY, ""); err != nil {
			return nil, err
		}
	}
	result.CheckGameOver(r)
	if r.Status != StatusFinished {
		r.advanceTurn()
	}
	return result, nil
}

func (r *Room) advanceTurn() {
	r.Turn = r.opponentCamp(r.Turn)
	r.Step++
}

func (r *Room) HasMovablePieces(camp string) bool {
	for _, piece := range r.Pieces {
		if piece.Alive && piece.Camp == camp && piece.Flipped {
			if piece.Type != PieceFlag && piece.Type != PieceMine {
				return true
			}
		}
	}
	return false
}

type BattleResult struct {
	AttackerID       string
	DefenderID       string
	AttackerType     string
	DefenderType     string
	AttackerAlive    bool
	DefenderAlive    bool
	Result           string
	ResultingPieceID string
	Winner           string
	Reason           string
}

func ResolveBattle(attacker, defender *Piece) *BattleResult {
	result := &BattleResult{
		AttackerID:    attacker.ID,
		DefenderID:    defender.ID,
		AttackerType:  attacker.Type,
		DefenderType:  defender.Type,
		AttackerAlive: true,
		DefenderAlive: true,
		Result:        "",
	}
	if defender.Type == PieceFlag {
		defender.Alive = false
		result.DefenderAlive = false
		result.Result = "attacker_win"
		result.Winner = attacker.Camp
		result.Reason = "flag_captured"
		return result
	}
	if attacker.Type == PieceBomb || defender.Type == PieceBomb {
		attacker.Alive = false
		defender.Alive = false
		result.AttackerAlive = false
		result.DefenderAlive = false
		result.Result = "both_die"
		return result
	}
	if defender.Type == PieceMine {
		if attacker.Type == PieceEngineer {
			defender.Alive = false
			result.DefenderAlive = false
			result.Result = "attacker_win"
			return result
		}
		attacker.Alive = false
		result.AttackerAlive = false
		result.Result = "defender_win"
		return result
	}
	if attacker.Rank > defender.Rank {
		defender.Alive = false
		result.DefenderAlive = false
		result.Result = "attacker_win"
		return result
	}
	if attacker.Rank < defender.Rank {
		attacker.Alive = false
		result.AttackerAlive = false
		result.Result = "defender_win"
		return result
	}
	attacker.Alive = false
	defender.Alive = false
	result.AttackerAlive = false
	result.DefenderAlive = false
	result.Result = "both_die"
	return result
}

func (r *BattleResult) CheckGameOver(room *Room) {
	if r.Winner != "" {
		room.Winner = r.Winner
		room.Reason = r.Reason
		room.Status = StatusFinished
		return
	}
	if !room.HasMovablePieces(CampRed) {
		room.Winner = CampBlue
		room.Status = StatusFinished
		room.Reason = "no_movable_pieces"
		return
	}
	if !room.HasMovablePieces(CampBlue) {
		room.Winner = CampRed
		room.Status = StatusFinished
		room.Reason = "no_movable_pieces"
		return
	}
}

func (r *Room) SyncData() map[string]any {
	boardView := make([][]map[string]any, r.Board.Rows)
	for y := 0; y < r.Board.Rows; y++ {
		boardView[y] = make([]map[string]any, r.Board.Cols)
		for x := 0; x < r.Board.Cols; x++ {
			cell := r.Board.Cells[y][x]
			if cell.PieceID == "" {
				continue
			}
			piece := r.Pieces[cell.PieceID]
			if piece == nil {
				continue
			}
			entry := map[string]any{
				"id":      piece.ID,
				"flipped": piece.Flipped,
			}
			if piece.Flipped {
				entry["type"] = piece.Type
				entry["camp"] = piece.Camp
			}
			boardView[y][x] = entry
		}
	}
	return map[string]any{
		"board": boardView,
		"turn":  r.Turn,
		"step":  r.Step,
	}
}

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

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func nowUnix() int64 {
	return time.Now().Unix()
}
