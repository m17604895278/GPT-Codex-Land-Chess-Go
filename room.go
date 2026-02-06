package game

import (
	"errors"
	"fmt"
)

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
