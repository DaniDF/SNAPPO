package engine

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"strconv"

	"main/logging"
)

const (
	EMPTY = uint8(0)
	SNAKE = uint8(1)
	APPLE = uint8(2)

	UP    = uint8(0)
	DOWN  = uint8(1)
	LEFT  = uint8(2)
	RIGHT = uint8(3)
)

type Cell struct {
	X uint8
	Y uint8
}

func (cell Cell) equals(c Cell) bool {
	return cell.X == c.X && cell.Y == c.Y
}

func (cell Cell) String() string {
	return "(" + strconv.Itoa(int(cell.X)) + "," + strconv.Itoa(int(cell.Y)) + ")"
}

type Board struct {
	ctx context.Context

	grid  [][]uint8
	apple Cell
	snake []Cell

	xSize uint8
	ySize uint8
}

func Init(ctx context.Context, xSize uint8, ySize uint8) Board {
	grid := make([][]uint8, xSize)

	for x := range xSize {
		grid[x] = make([]uint8, ySize)

		for y := range ySize {
			grid[x][y] = EMPTY
		}
	}

	snake := []Cell{{X: uint8(rand.IntN(int(xSize))), Y: uint8(rand.IntN(int(ySize)))}, {X: 0, Y: 0}}
	snake[1].X = snake[0].X
	snake[1].Y = snake[0].Y

	for !isSnakeValid(snake, xSize, ySize) {
		snake[1].X = snake[0].X + uint8(rand.IntN(2)-1)
		snake[1].Y = snake[0].Y + uint8(rand.IntN(2)-1)
	}

	apple := generateNewApple(snake, xSize, ySize)

	grid[apple.X][apple.Y] = APPLE
	grid[snake[0].X][snake[0].Y] = SNAKE
	grid[snake[1].X][snake[1].Y] = SNAKE

	return Board{
		ctx: ctx,

		grid:  grid,
		apple: apple,
		snake: snake,

		xSize: xSize,
		ySize: ySize,
	}
}

func (board Board) GetXSize() uint8 {
	return board.xSize
}

func (board Board) GetYSize() uint8 {
	return board.ySize
}

func (board Board) GetApple() Cell {
	return board.apple
}

func (board Board) GetSnake() []Cell {
	return slices.Clone(board.snake) // TODO: this crates a copy of the object but for efficiency consider to return the pointer (unwanted modification)
}

func (board Board) GetGrid() [][]uint8 {
	return board.grid
}

func (board Board) Move(direction uint8) error {
	log := board.ctx.Value("logger").(logging.Logger)

	shiftCell := Cell{X: 0, Y: 0}

	switch direction {
	case UP:
		shiftCell = Cell{X: board.snake[0].X, Y: board.snake[0].Y + 1}

	case DOWN:
		if int8(board.snake[0].Y)-1 < 0 {
			log.Error("[Board] Snake moving out of the board: (" + strconv.Itoa(int(board.snake[0].X)) + "," + strconv.Itoa(int(board.snake[0].Y)-1) + ")")
			return errors.New("New cell out of bound")
		}

		shiftCell = Cell{X: board.snake[0].X, Y: board.snake[0].Y - 1}

	case LEFT:
		if int8(board.snake[0].X)-1 < 0 {
			log.Error("[Board] Snake moving out of the board: (" + strconv.Itoa(int(board.snake[0].X)-1) + "," + strconv.Itoa(int(board.snake[0].Y)) + ")")
			return errors.New("New cell out of bound")
		}

		shiftCell = Cell{X: board.snake[0].X - 1, Y: board.snake[0].Y}

	case RIGHT:
		shiftCell = Cell{X: board.snake[0].X + 1, Y: board.snake[0].Y}

	default:
		log.Error("[Board] Unsupported move: " + strconv.Itoa(int(direction)))
		return errors.New("Invalid move")
	}

	err := board.validateMove(shiftCell)
	if err != nil {
		log.Error("[Board] Next position not valid: " + shiftCell.String() + " - " + err.Error())
		return err
	}

	board.shiftSnakeToCell(shiftCell)

	return nil
}

func (board Board) validateMove(cell Cell) error {
	if cell.X >= board.xSize || cell.Y >= board.ySize {
		return errors.New("New cell out of bound")
	} else if slices.Contains(board.snake, cell) {
		return errors.New("New cell is cannibalism")
	}

	return nil
}

func (board Board) shiftSnakeToCell(cell Cell) {
	if cell.equals(board.apple) {
		board.snake = append(board.snake, board.snake[len(board.snake)-1])
		board.apple = generateNewApple(board.snake, board.xSize, board.ySize)
	}

	shiftPosition := cell
	for i := range len(board.snake) {
		temp := board.snake[i]

		board.snake[i].X = shiftPosition.X
		board.snake[i].Y = shiftPosition.Y
		board.grid[shiftPosition.X][shiftPosition.Y] = SNAKE

		shiftPosition = temp
		board.grid[shiftPosition.X][shiftPosition.Y] = EMPTY
	}
}

func generateNewApple(snake []Cell, xSize uint8, ySize uint8) Cell { // TODO Choose only among valid cells
	apple := Cell{X: snake[0].X, Y: snake[0].Y}

	for slices.Contains(snake, apple) {
		apple.X = uint8(rand.IntN(int(xSize)))
		apple.Y = uint8(rand.IntN(int(ySize)))
	}

	return apple
}

func isSnakeValid(snake []Cell, xSize uint8, ySize uint8) bool {
	result := true

	for i := range len(snake) - 1 {
		result = result &&
			!snake[i+1].equals(snake[i]) &&
			snake[i+1].X < xSize && snake[i+1].Y < ySize &&
			((snake[i].X == snake[i+1].X) || (snake[i].Y == snake[i+1].Y)) // The snake can only make 90° turns
	}

	return result
}
