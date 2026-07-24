package tui

import (
	"fmt"
	"game/engine"
	"strings"
)

const (
	// ANSI Escape Codes: Bold Red text and Reset code
	red    = "\033[1;31m"
	green  = "\033[1;32m"
	yellow = "\033[1;33m"
	reset  = "\033[0m"
)

// Renders a given board on stdout
func RenderBoard(board engine.Board) {
	fmt.Print("\033[H")

	fmt.Println("+" + strings.Repeat("--", int(board.XSize)) + "+")

	for revY := range board.YSize {
		y := board.YSize - 1 - revY // Flipped the Y axis
		fmt.Print("|")

		for x := range board.XSize {
			switch board.Grid[x][y] {
			case engine.APPLE:
				fmt.Print(yellow + "★" + reset + ".")

			case engine.SNAKE:
				if board.Snake[0].X == x && board.Snake[0].Y == y {
					fmt.Print(green + "⏺" + reset + ".")
				} else {
					fmt.Print(green + "⏹" + reset + ".")
				}
			default:
				fmt.Print(" .")
			}
		}

		fmt.Println("|")
	}

	fmt.Println("+" + strings.Repeat("--", int(board.XSize)) + "+")
}

// Clears the output
func ClearBoard() {
	fmt.Print("\033[H\033[2J")
}

func RenderGameOver() {
	banner := `
 ██████╗  █████╗ ███╗   ███╗███████╗    ██████╗ ██╗   ██╗███████╗██████╗ 
██╔════╝ ██╔══██╗████╗ ████║██╔════╝   ██╔═══██╗██║   ██║██╔════╝██╔══██╗
██║  ███╗███████║██╔████╔██║█████╗     ██║   ██║██║   ██║█████╗  ██████╔╝
██║   ██║██╔══██║██║╚██╔╝██║██╔══╝     ██║   ██║██║   ██║██╔══╝  ██╔══██╗
╚██████╔╝██║  ██║██║ ╚═╝ ██║███████╗   ╚██████╔╝╚██████╔╝███████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝    ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝
`
	fmt.Print(red + banner + reset)
}
