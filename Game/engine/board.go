package engine

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"game/logging"
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
	X uint8 `json:"x"`
	Y uint8 `json:"y"`
}

// Compares this Cell with c. Returns true if both X and Y are equal.
func (cell Cell) equals(c Cell) bool {
	return cell.X == c.X && cell.Y == c.Y
}

// Generates a string version of this cell.
func (cell Cell) String() string {
	return "(" + strconv.Itoa(int(cell.X)) + "," + strconv.Itoa(int(cell.Y)) + ")"
}

type Board struct {
	ctx context.Context `json:"-"`

	Grid  [][]uint8 `json:"grid"`
	Apple Cell      `json:"apple"`
	Snake []Cell    `json:"snake"`

	XSize    uint8 `json:"xMax"`
	YSize    uint8 `json:"yMax"`
	Gameover bool  `json:"gameover"`
}

// Initialise a new Board given X and Y sizes.
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

		Grid:  grid,
		Apple: apple,
		Snake: snake,

		XSize:    xSize,
		YSize:    ySize,
		Gameover: false,
	}
}

func (board Board) getGrid() [][]uint8 {
	if board.Grid == nil {
		return nil
	}

	var result = make([][]uint8, len(board.Grid))
	for i := range board.Grid {
		result[i] = slices.Clone(board.Grid[i])
	}

	return result
}

func (board Board) getApple() Cell {
	return board.Apple
}

func (board Board) getSnake() []Cell {
	return slices.Clone(board.Snake)
}

func (board Board) getXSize() uint8 {
	return board.XSize
}

func (board Board) getYSize() uint8 {
	return board.YSize
}

func (board Board) getGameover() bool {
	return board.Gameover
}

// Geerates a string version representing the snake.
func (board Board) GetSnakeString() string {
	var result strings.Builder

	for _, cell := range board.Snake {
		result.WriteString(cell.String())
	}

	return result.String()
}

func (board Board) Clone() Board {
	return Board{
		ctx: nil,

		Grid:  board.getGrid(),
		Apple: board.getApple(),
		Snake: board.getSnake(),

		XSize:    board.getXSize(),
		YSize:    board.getYSize(),
		Gameover: board.getGameover(),
	}
}

/*
Move snake in the specified direction.
Allowed directions: engine.UP, engine.DOWN, engine.LEFT, engine.RIGHT.
Returns an error is the new cell of the snake's head is: out-of-bound, the snake itself (cannibalism).
*/
func (board *Board) Move(direction uint8) error {
	log := board.ctx.Value("logger").(logging.Logger)

	shiftCell := Cell{X: 0, Y: 0}

	switch direction {
	case UP:
		shiftCell = Cell{X: board.Snake[0].X, Y: board.Snake[0].Y + 1}

	case DOWN:
		if int(board.Snake[0].Y)-1 < 0 {
			board.Gameover = true
			log.Error("[Board] Snake moving out of the board: (" + strconv.Itoa(int(board.Snake[0].X)) + "," + strconv.Itoa(int(board.Snake[0].Y)-1) + ")")
			return errors.New("New cell out of bound")
		}

		shiftCell = Cell{X: board.Snake[0].X, Y: board.Snake[0].Y - 1}

	case LEFT:
		if int(board.Snake[0].X)-1 < 0 {
			board.Gameover = true
			log.Error("[Board] Snake moving out of the board: (" + strconv.Itoa(int(board.Snake[0].X)-1) + "," + strconv.Itoa(int(board.Snake[0].Y)) + ")")
			return errors.New("New cell out of bound")
		}

		shiftCell = Cell{X: board.Snake[0].X - 1, Y: board.Snake[0].Y}

	case RIGHT:
		shiftCell = Cell{X: board.Snake[0].X + 1, Y: board.Snake[0].Y}

	default:
		log.Error("[Board] Unsupported move: " + strconv.Itoa(int(direction)))
		return errors.New("Invalid move")
	}

	err := board.validateMove(shiftCell)
	if err != nil {
		board.Gameover = true
		log.Error("[Board] Next position not valid: " + shiftCell.String() + " - " + err.Error())
		return err
	}

	board.shiftSnakeToCell(shiftCell)

	return nil
}

// Checks if moving the snake's head into the defined cell is valid or not
func (board Board) validateMove(cell Cell) error {
	if board.Gameover {
		return errors.New("Game is already over")
	} else if cell.X >= board.XSize || cell.Y >= board.YSize {
		return errors.New("New cell out of bound")
	} else if slices.Contains(board.Snake, cell) {
		return errors.New("New cell is cannibalism")
	}

	return nil
}

// Moves the snake's head into the provided cell and the rest of the snake accordingly
func (board *Board) shiftSnakeToCell(cell Cell) {
	log := board.ctx.Value("logger").(logging.Logger)

	if cell.equals(board.Apple) {
		board.Snake = append(board.Snake, board.Snake[len(board.Snake)-1])
		log.Debug("[Board] Snake ate an apple, new snake length " + strconv.Itoa(len(board.Snake)))

		board.Apple = generateNewApple(board.Snake, board.XSize, board.XSize)
		board.Grid[board.Apple.X][board.Apple.Y] = APPLE
		log.Debug("[Board] Generated new apple at " + board.Apple.String())
	}

	shiftPosition := cell
	for i := range len(board.Snake) {
		temp := board.Snake[i]

		board.Snake[i].X = shiftPosition.X
		board.Snake[i].Y = shiftPosition.Y
		board.Grid[shiftPosition.X][shiftPosition.Y] = SNAKE

		shiftPosition = temp
		board.Grid[shiftPosition.X][shiftPosition.Y] = EMPTY
	}
}

// Generates a new valid apple given the grid size and the snake position
func generateNewApple(snake []Cell, xSize uint8, ySize uint8) Cell { // TODO Choose only among valid cells
	apple := Cell{X: snake[0].X, Y: snake[0].Y}

	for slices.Contains(snake, apple) {
		apple.X = uint8(rand.IntN(int(xSize)))
		apple.Y = uint8(rand.IntN(int(ySize)))
	}

	return apple
}

/*
Checks if the provided snake is valid.
Eg. No invalid turns, no 45° turns
*/
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
