package tui

import (
	"fmt"
	"main/engine"
	"strings"
)

func RenderBoard(board engine.Board) {
	fmt.Print("\033[H")

	grid := board.GetGrid()

	fmt.Println("+" + strings.Repeat("--", int(board.GetXSize())) + "+")

	for revY := range board.GetYSize() {
		y := board.GetYSize() - 1 - revY // Flipped the Y axis
		fmt.Print("|")

		for x := range board.GetXSize() {
			switch grid[x][y] {
			case engine.APPLE:
				fmt.Print("★.")

			case engine.SNAKE:
				if board.GetSnake()[0].X == x && board.GetSnake()[0].Y == y {
					fmt.Print("⏺.")
				} else {
					fmt.Print("⏹.")
				}
			default:
				fmt.Print(" .")
			}
		}

		fmt.Println("|")
	}

	fmt.Println("+" + strings.Repeat("--", int(board.GetXSize())) + "+")
}
