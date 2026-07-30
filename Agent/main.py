from concurrent.futures import ProcessPoolExecutor, as_completed

import logging
import socket

import typer
from typing_extensions import Annotated
import math
import os

import board
import logger

from agent import Agent
from actor import ActorCritic

import utils

DEFAULT_SOCKET = "/tmp/snake.sock"
NUM_GAMES = 10
NUM_ITERATIONS = 100
X_GAME_SIZE = 20
Y_GAME_SIZE = 20

launcher = typer.Typer(help="Play the game")

@launcher.command()
def main(
    socket_path: Annotated[
            str, typer.Option("--socket", "-s", help=f"Path to the socket (default: {DEFAULT_SOCKET})")
    ] = DEFAULT_SOCKET,
    debug_enabled: Annotated[
            bool, typer.Option("-d", help="Enables debug logging")
        ] = False,
    ckpt_dir: Annotated[
            str, typer.Option("-c", help="Specify the path where to save checkpoints, implicitly enables this feature (disabled otherwise)")
        ] = "",
    load_ckpt_enabled: Annotated[
            bool, typer.Option("-C", help="Enables the loading of the model and weights at startup. -c must be specified")
        ] = False,
    x_size: Annotated[
            int, typer.Option("-x", help=f"Sets the width of the board (default: {X_GAME_SIZE})")
        ] = X_GAME_SIZE,
    y_size: Annotated[
            int, typer.Option("-y", help=f"Sets the height of the board (default: {Y_GAME_SIZE})")
        ] = Y_GAME_SIZE,
    n_games: Annotated[
                int, typer.Option("-g", help=f"Set the number of games in each iteraction (default: {NUM_GAMES})")
            ] = NUM_GAMES,
    n_iterations: Annotated[
                int, typer.Option("-i", help=f"Set the number of iteractions (default: {NUM_ITERATIONS})")
            ] = NUM_ITERATIONS,
):
    log = None
    if debug_enabled:
        log = logger.init_debug()
    else:
        log = logger.init()

    flag_save_policy = len(ckpt_dir) > 0

    policy = ActorCritic(input_shape=(x_size*y_size,), action_dim=4)

    if load_ckpt_enabled and len(ckpt_dir) == 0:
        log.error("[Main] Unspecified -c checkpoit dir when enabled -C option (loading model)")
        os._exit(-1)
    elif load_ckpt_enabled:
        utils.load_policy(policy, dir=ckpt_dir, only_weights=False)
    elif flag_save_policy:
        utils.save_policy(policy, dir=ckpt_dir, only_weights=False)

    agent = Agent(policy=policy, logger=log)


    #with ProcessPoolExecutor() as executor:
    #    tasks = [executor.submit(play_game, log, socket_path, policy) for _ in range(NUM_ACTORS)]

    #    for future in as_completed(tasks):
    #        try:
    #            memories.append(future.result())
    #        except Exception as e:
    #            log.error(f"[Main] Task completed with exception: {e}")

    for iteration in range(n_iterations):
        for _ in range(n_games):
            play_game(log, socket_path, agent, x_size, y_size)
    
        losses = agent.train()
        log.info(f"[Main] Training iteration {iteration} losses: {losses}")
        agent.clear_memory()

        if len(ckpt_dir) > 0:
            actor_dir_path = os.path.join(ckpt_dir,"actor")
            critic_dir_path = os.path.join(ckpt_dir,"critic")

            os.makedirs(actor_dir_path, exist_ok=True)
            os.makedirs(critic_dir_path, exist_ok=True)

            if iteration%10 == 9 and len(ckpt_dir) > 0:
                if flag_save_policy:
                    utils.save_policy(policy, dir=ckpt_dir)

    if flag_save_policy:
        utils.save_policy(policy, dir=ckpt_dir)


def play_game(log: logging.Logger, socket_path: str, agent: Agent, x_size: int, y_size: int):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
        conn.connect(socket_path)

        conn.sendall(("{\"xMax\":" + f"{x_size}" + ",\"yMax\":" + f"{y_size}" + "}").encode())

        old_game_score = 0

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
                    
                    snake_head = game.snake[0]
                    apple_distance = math.sqrt(math.pow(snake_head.x - game.apple.x, 2) + math.pow(snake_head.y - game.apple.y, 2))

                    if game.get_score() == old_game_score:
                        agent.reward(-apple_distance)
                    else:
                        score = game.get_score()
                        agent.reward(score - old_game_score)
                        old_game_score = score
                    

                move = agent.predict(game)
                log.debug(f"[Main] Chose move: {move.name}")
                conn.sendall(move.marshal().encode())
            else:
                flag_stop = True
                agent.reward(game.get_score(), is_terminal=True)
                log.info("[Main] GAME OVER")


if __name__ == "__main__":
    launcher()