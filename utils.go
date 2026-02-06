package game

import "time"

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func nowUnix() int64 {
	return time.Now().Unix()
}
