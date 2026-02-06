package game

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
