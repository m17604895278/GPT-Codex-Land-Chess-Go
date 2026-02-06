package game

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
