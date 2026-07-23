package main

import (
	"context"
	"encoding/json"
	"errors"
	"game/engine"
	"game/logging"
	tui "game/ui/TUI"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/alexflint/go-arg"
)

const (
	DEFAULT_SOCKET = "/tmp/snake.sock"
)

type Args struct {
	XSize uint8 `arg:"-x" default:"20" help:"Sets the width of the board"`
	YSize uint8 `arg:"-y" default:"20" help:"Sets the heigh of the board"`

	UI bool `arg:"--ui" default:"false" help:"Enable the terminal UI"`

	socket string `arg:"--socket" default:"" help:"Define custom socket"`

	DebugEnabled bool `arg:"-d,--debug" default:"false" help:"Enable debug logging"`
}

func main() {
	var args Args
	arg.MustParse(&args)

	debugLevel := slog.LevelInfo // TODO
	if args.DebugEnabled {
		debugLevel = slog.LevelDebug
	}

	socketPath := DEFAULT_SOCKET
	if len(args.socket) > 0 {
		socketPath = args.socket
	}

	ctx := context.Background()
	ctx, log := logging.Init(ctx, debugLevel)

	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		log.Error("[Main] Error while removing old socket: " + err.Error())
		os.Exit(-1)
	}

	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Error("[Main] Error while creating the socket: " + err.Error())
		os.Exit(-2)
	}
	defer sock.Close()
	defer os.Remove(socketPath)

	log.Debug("[Main] Listening on socket " + socketPath)

	for {
		conn, err := sock.Accept()
		go func() {
			if err != nil {
				log.Warn("[Main] Error while accepting a new connection: " + err.Error())
				return
			}
			defer conn.Close()
			if err := handleConnection(ctx, conn, args.XSize, args.YSize, args.UI); err != nil {
				log.Error("[Main] Game ended with error: " + err.Error())
				return
			}
		}()
	}
}

func handleConnection(ctx context.Context, conn net.Conn, xSize uint8, ySize uint8, enableUI bool) error {
	log := ctx.Value("logger").(logging.Logger)

	log.Info("[Main-conn] Game started (" + strconv.Itoa(int(xSize)) + "," + strconv.Itoa(int(ySize)) + ")")

	game := engine.Init(ctx, xSize, ySize)

	for !game.Gameover {
		if enableUI {
			tui.RenderBoard(game)
		}
		log.Debug("[Main-conn] Snake in " + game.GetSnakeString())

		gameJ, err := json.Marshal(game)
		if err != nil {
			log.Warn("[Main-conn] Fatal error while marshaling game: " + err.Error())
			return err
		}

		_, err = conn.Write([]byte(gameJ))
		if err != nil {
			log.Warn("[Main-conn] Fatal error while writing data to socket: " + err.Error())
			return err
		}

		buffer := make([]byte, 256)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Warn("[Main-conn] Fatal error while reading from socket: " + err.Error())
			return err
		}
		var moveObj map[string]string
		err = json.Unmarshal(buffer[:n], &moveObj)
		if err != nil {
			log.Warn("[Main-conn] Fatal error while unmashaling data from socket: " + err.Error())
			return err
		}

		directionS, present := moveObj["direction"]
		if present {
			direction, err := parseMoveDirection(directionS)
			if err == nil {
				err := game.Move(direction)
				if err != nil {
					log.Debug("[Main-conn] Invalid move: " + err.Error())
				}
			}
		}

		if game.Gameover {
			log.Info("[Main-conn] GAME OVER")

			gameJ, err := json.Marshal(game)
			if err != nil {
				log.Warn("[Main-conn] Error while marshaling game (over): " + err.Error())
				return err
			}

			_, err = conn.Write([]byte(gameJ))
			if err != nil {
				log.Warn("[Main-conn] Error while writing data to socket (gameover): " + err.Error())
				return err
			}
		}
	}

	return nil
}

func parseMoveDirection(str string) (uint8, error) {
	var result uint8

	switch str {
	case "UP":
		result = engine.UP
	case "DOWN":
		result = engine.DOWN
	case "LEFT":
		result = engine.LEFT
	case "RIGHT":
		result = engine.RIGHT
	default:
		return 0, errors.New("Unsupported direction")
	}

	return result, nil
}
