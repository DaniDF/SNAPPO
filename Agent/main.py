import socket

import typer
from typing_extensions import Annotated

import board
import logger

DEFAULT_SOCKET = "/tmp/snake.sock"

launcher = typer.Typer(help="Play the game")

@launcher.command()
def main(
    socket_path: Annotated[
            str, typer.Option("--socket", "-s", help="Path to the socket")
    ] = DEFAULT_SOCKET,
    debug_enabled: Annotated[
            bool, typer.Option("-d", help="Enables debug logging")
        ] = False
):
    log = None
    if debug_enabled:
        log = logger.init_debug()
    else:
        log = logger.init()
    
    
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
        conn.connect(socket_path)

        flag_stop = False

        while not flag_stop:
            data = conn.recv(4096)

            game = board.Unmarshal(data)

            if not game.gameover:
                move = think(game)
                log.debug(f"[Main] Chose move: {move.name}")
                conn.sendall(move.marshal().encode())
            else:
                flag_stop = True
                log.info("[Main] GAME OVER")


def think(game: board.Board) -> board.Move:
    return board.Move.LEFT # TODO


if __name__ == "__main__":
    launcher()