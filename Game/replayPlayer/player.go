package main

import (
	"encoding/json"
	"fmt"
	"game/history"
	tui "game/ui/TUI"
	"os"
	"time"

	"github.com/alexflint/go-arg"
)

type Args struct {
	HistoryFile string `arg:"-H,--history-file" help:"Specify the history file to replay"`

	FastRender bool `arg:"-F,--fast-render" default:"false" help:"Enable fast rendering"`
}

func main() {
	var args Args
	arg.MustParse(&args)

	renderTime := 500
	if args.FastRender {
		renderTime = 150
	}

	fileData, err := os.ReadFile(args.HistoryFile)
	var history history.History
	err = json.Unmarshal(fileData, &history)
	if err != nil {
		fmt.Println("[Player] Error while unmarshaling history file data: " + err.Error())
		os.Exit(-1)
	}

	tui.ClearBoard()

	for _, record := range history.Records {
		tui.RenderBoard(record)
		time.Sleep(time.Duration(renderTime) * time.Millisecond)
	}
}
