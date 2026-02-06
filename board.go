package game

import "fmt"

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
