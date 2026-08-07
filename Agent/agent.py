import numpy as np

import board as b
import logging
from actor import ActorCritic

class Agent:
    def __init__(self, policy: ActorCritic, logger: logging.Logger|None = None):
        self.__policy = policy

        self.clear_memory()

        self.__log = logger

    def clear_memory(self):
        self.__memory = {
            "states": [],
            "actions": [],
            "values": [],
            "log_probs": [],
            "rewards": [],
            "is_terminals": []
        }

    def predict(self, board: b.Board) -> b.Move:
        state = np.array(board.grid, dtype=np.float32).ravel()
        state /= np.linalg.norm(state)
        action, log_prob, value = self.__policy.get_action(state=state)
    
        self.__memory["states"].append(state)
        self.__memory["actions"].append(action)
        self.__memory["values"].append(value)
        self.__memory["log_probs"].append(log_prob)

        self.__log.debug(f"[Agent] Action: {action} Log_prob: {log_prob}")

        return b.from_int_to_Move(action)

    def reward(self, reward: float, is_terminal: bool = False):
        self.__memory["rewards"].append(reward)
        self.__memory["is_terminals"].append(is_terminal)


    def train(self):
        return self.__policy.ppo_update(self.__memory)