from enum import Enum
import json
import base64

class Cell:
    x: int
    y: int

    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y

    def String(self) -> str:
        return f"({self.x},{self.y})"

class Board:
    def __init__(self):
        self.grid: list[list[int]] = []
        self.apple: Cell = None
        self.snake: list[Cell] = []
    
        self.xSize: int = None
        self.ySize: int = None
        self.gameover: bool = None

    def String(self) -> str:
        result = ""
        for row in self.grid:
            result += "["

            for col in row:
                result += str(col) + ","

            result += "]\n"

        result += f"apple: {self.apple.String()}\n"

        result += "snake:"
        for snake_cell in self.snake:
            result += f" {snake_cell.String()}"

        return result + f"\nxSize: {self.xSize}\nySize: {self.ySize}\ngameover: {self.gameover}"

    def get_score(self) -> int:
        return len(self.snake)


def Unmarshal(s: str) -> Board:
    data = json.loads(s)

    result = Board()

    for row in data["grid"]:
        result.grid.append(list(base64.b64decode(row)))

    result.apple = Cell(x=data["apple"]["x"], y=data["apple"]["y"])

    for snale_cell in data["snake"]:
        result.snake.append(Cell(x=snale_cell["x"], y=snale_cell["y"]))

    result.xSize = data["xMax"]
    result.ySize = data["yMax"]
    result.gameover = data["gameover"]

    return result

class Move(Enum):
    UP = 0
    DOWN = 1
    LEFT = 2
    RIGHT = 3

    def marshal(self) -> str:
        return "{\"direction\": \"" + self.name + "\"}"

def from_int_to_Move(value: int) -> Move:
    result = None

    if value == 0:
        result = Move.UP
    elif value == 1:
        result = Move.DOWN
    elif value == 2:
        result = Move.LEFT
    elif value == 3:
        result = Move.RIGHT
    else:
        raise Exception("Invadid enum value")
    
    return result