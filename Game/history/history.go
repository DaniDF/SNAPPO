package history

import (
	"game/engine"
	"strconv"
	"time"
)

type History struct {
	StartTime int64 `json:"start"`
	EndTime   int64 `json:"finish"`

	Records []engine.Board `json:"records"`
	Moves   []uint8        `json:"moves"`
	Score   uint           `json:"score"`
}

func StartHistory(initialState engine.Board) History {
	return History{
		StartTime: time.Now().UnixMicro(),
		EndTime:   -1,

		Records: []engine.Board{initialState.Clone()},
		Moves:   []uint8{},
		Score:   0,
	}
}

func (history *History) AddRecord(move uint8, record engine.Board) {
	history.Records = append(history.Records, record.Clone())
	history.Moves = append(history.Moves, move)
	history.Score = uint(len(record.Snake))
}

func (history *History) Finish() {
	history.EndTime = time.Now().UnixMicro()
}

func (history History) String() string {
	result := "Game started at " + time.UnixMicro(history.StartTime).Format(time.RFC3339Nano) + ", "

	if history.EndTime < 0 {
		result += "yet to finish"
	} else {
		result += "finished at " + time.UnixMicro(history.EndTime).Format(time.RFC3339Nano)
	}

	result += ", score: " + strconv.Itoa(int(history.Score))
	result += ", moves: " + strconv.Itoa(len(history.Moves))

	return result
}
