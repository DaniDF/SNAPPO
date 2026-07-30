package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"game/engine"
	"game/interactivePlayer/utils"
	tui "game/ui/TUI"

	"github.com/alexflint/go-arg"
)

const (
	DEFAULT_SOCKET = "/tmp/snake.sock"
)

type Args struct {
	XSize uint8 `arg:"-x" default:"20" help:"Sets the width of the board"`
	YSize uint8 `arg:"-y" default:"20" help:"Sets the heigh of the board"`

	Socket string `arg:"--socket" default:"" help:"Define custom socket"`
}

func main() {
	var args Args
	arg.MustParse(&args)

	socketPath := DEFAULT_SOCKET
	if len(args.Socket) > 0 {
		socketPath = args.Socket
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("[Main] Error while dialing the socket: " + err.Error())
		os.Exit(-2)
	}
	defer conn.Close()

	gameConfigsStr := fmt.Sprintf("{\"xMax\":%s,\"yMax\":%s}", strconv.Itoa(int(args.XSize)), strconv.Itoa(int(args.YSize)))
	_, err = conn.Write([]byte(gameConfigsStr))
	if err != nil {
		fmt.Println("[Main] Error while sending game configs")
		os.Exit(-3)
	}

	tui.ClearBoard()

	flagStop := false
	buffer := make([]byte, 4096)
	var game engine.Board
	for !flagStop {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[Player] Error while reading from socket")
		}

		err = json.Unmarshal(buffer[:n], &game)
		if err != nil {
			fmt.Println("[Player] Error while unmashaling data from socket: " + err.Error())
			return
		}

		if game.Gameover {
			flagStop = true
			tui.ClearBoard()
			tui.RenderGameOver()

		} else {
			tui.RenderBoard(game)

			flagKeyValid := false
			var move string
			for !flagKeyValid {
				key, err := utils.ReadKey()
				if err == nil {
					flagKeyValid = true

					switch string(key) {
					case "w":
						move = "UP"
					case "s":
						move = "DOWN"
					case "a":
						move = "LEFT"
					case "d":
						move = "RIGHT"
					default:
						flagKeyValid = false
					}
				}
			}

			_, err = conn.Write([]byte("{\"direction\": \"" + move + "\"}"))
			if err != nil {
				fmt.Println("[Player] Error while sending direction to socket + " + err.Error())
				return
			}
		}
	}

}
