from concurrent.futures import ProcessPoolExecutor, as_completed

import logging
import socket

import typer
from typing_extensions import Annotated

import board
import logger

from agent import Agent
from actor import ActorCritic

DEFAULT_SOCKET = "/tmp/snake.sock"
NUM_GAMES = 10
NUM_ITERATION = 200
GAME_SIZE = 20*20

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

    policy = ActorCritic(input_shape=(GAME_SIZE,), action_dim=4)
    agent = Agent(policy=policy, logger=log)

    #with ProcessPoolExecutor() as executor:
    #    tasks = [executor.submit(play_game, log, socket_path, policy) for _ in range(NUM_ACTORS)]

    #    for future in as_completed(tasks):
    #        try:
    #            memories.append(future.result())
    #        except Exception as e:
    #            log.error(f"[Main] Task completed with exception: {e}")

    for iteration in range(NUM_ITERATION):
        for _ in range(NUM_GAMES):
            play_game(log, socket_path, agent)
    
        losses = agent.train()
        log.info(f"[Main] Training iteration {iteration} losses: {losses}")


def play_game(log: logging.Logger, socket_path: str, agent: Agent):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
        conn.connect(socket_path)

        flag_stop = False
        flag_first_board = True
        while not flag_stop:
            data = conn.recv(4096)

            game = board.Unmarshal(data)

            if not game.gameover:
                if flag_first_board:
                    flag_first_board = False
                else:
                    # TODO Think about giving rewards based on score difference instead of the current score (current_score - previews_score) to actually reward the model when it gets an apple
                    # Possible problems: getting an apple when the snake is 2 is the same as when it is 50, but the difficulty is much higher (future me problem)
                    agent.reward(game.get_score())

                move = agent.predict(game)
                log.debug(f"[Main] Chose move: {move.name}")
                conn.sendall(move.marshal().encode())
            else:
                flag_stop = True
                agent.reward(game.get_score(), is_terminal=True)
                log.info("[Main] GAME OVER")


if __name__ == "__main__":
    launcher()