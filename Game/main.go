package main

import (
	"context"
	"log/slog"
	tui "main/TUI"
	"main/engine"
	"main/logging"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/alexflint/go-arg"
)

type Args struct {
	XSize uint8 `arg:"-x" default:"20" help:"Sets the width of the board"`
	YSize uint8 `arg:"-y" default:"20" help:"Sets the heigh of the board"`

	DebugEnabled bool `arg:"-d,--debug" default:"false" help:"Enable debug logging"`
}

func main() {
	var args Args
	arg.MustParse(&args)

	debugLevel := slog.LevelInfo // TODO
	if args.DebugEnabled {
		debugLevel = slog.LevelDebug
	}

	ctx := context.Background()
	ctx, log := logging.Init(ctx, debugLevel)

	log.Info("[Main] Game started (" + strconv.Itoa(int(args.XSize)) + "," + strconv.Itoa(int(args.YSize)) + ")")

	game := engine.Init(ctx, args.XSize, args.YSize)

	tui.RenderBoard(game)
	log.Debug("[Main] Snake in " + game.GetSnake()[0].String() + game.GetSnake()[1].String())

	flagStop := false
	for !flagStop {
		err := game.Move(chooseRandomMove(ctx))

		if err != nil {
			flagStop = true
			log.Info("[Main] GAME OVER")
			log.Debug("[Main] Game ended for invalid move")

		} else {
			time.Sleep(2 * time.Second)

			tui.RenderBoard(game)
			log.Debug("[Main] Snake in " + game.GetSnake()[0].String() + game.GetSnake()[1].String())
		}
	}
}

func chooseRandomMove(ctx context.Context) uint8 {
	log := ctx.Value("logger").(logging.Logger)

	move := uint8(rand.IntN(4))

	switch move {
	case engine.UP:
		log.Debug("[Main] Chose random move UP")
	case engine.DOWN:
		log.Debug("[Main] Chose random move DOWN")
	case engine.LEFT:
		log.Debug("[Main] Chose random move LEFT")
	case engine.RIGHT:
		log.Debug("[Main] Chose random move RIGHT")
	}

	return move
}
