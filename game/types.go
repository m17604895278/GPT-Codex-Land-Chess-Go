package game

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
