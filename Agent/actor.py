import tensorflow as tf
import tensorflow_probability as tfp
import numpy as np

def sum_proba_to_one(proba):
    return proba/proba.sum()

class ActorCritic(tf.keras.Model):
    def __init__(self, input_shape: tuple[int], action_dim: int, learning_rate = 3e-4):
        super(ActorCritic, self).__init__()
        
        # Defining the actor model
        self.__actor = tf.keras.models.Sequential([
            tf.keras.layers.Input(shape=input_shape, dtype=tf.float32),
            tf.keras.layers.Dense(64, activation="tanh", dtype=tf.float32),  # TODO Evaluate relu
            tf.keras.layers.Dense(64, activation="tanh", dtype=tf.float32),
            tf.keras.layers.Dense(action_dim, activation="softmax")
        ])
        self.__actor.compile(optimizer=tf.keras.optimizers.Adam(learning_rate))
        
        # Defining the critic model
        self.__critic = tf.keras.models.Sequential([
            tf.keras.layers.Input(shape=input_shape, dtype=tf.float32),
            tf.keras.layers.Dense(64, activation="tanh", dtype=tf.float32),  # TODO Evaluate relu
            tf.keras.layers.Dense(64, activation="tanh", dtype=tf.float32),
            tf.keras.layers.Dense(1)
        ])
        self.__critic.compile(optimizer=tf.keras.optimizers.Adam(learning_rate))

        
    def get_action(self, state):
        action_probs = self.__actor(tf.convert_to_tensor([state]), dtype=tf.float32)[0]
        dist = tfp.distributions.Categorical(action_probs)

        action = dist.sample()
        log_prob = dist.log_prob(action)
        value = self.evaluate(state)
        
        return action, log_prob, value
    
    
    def evaluate(self, state):
        return self.__critic(tf.convert_to_tensor([state]), dtype=tf.float32)[0]
    

    #def save_model_weights(self, filename):
    #        self.__model_policy__.save_weights(filename)
    
    #def load_model_weights(self, filename):
    #    self.__model_policy__.load_weights(filename)

    #def save_model(self, filename):
    #    self.__model_policy__.save(filename)

    #def load_model(self, filename):
    #    self.__model_policy__ = tf.keras.models.load_model(filename)

    def ppo_update(self, memories, batch_size=10, gamma=0.99, lambd=0.95, epsilon_clip=0.2, k_epochs=10):
        losses = []
        for _ in range(k_epochs):
            states = np.array(memories["states"])
            actions = np.array(memories["actions"])
            values = np.array(memories["values"])
            log_probs = np.array(memories["log_probs"])
            rewards = np.array(memories["rewards"])
            is_terminals = np.array(memories["is_terminals"])

            advantages = np.zeros(len(states), dtype=np.float32)
            for t in range(len(states)):
                reward_t = rewards[t]
                if is_terminals[t]:
                    reward_t = 0

                count = 0
                a_t = np.array([0], dtype=np.float32)
                for i in range(t, len(states)-1):
                    delta_t = reward_t + gamma*values[i+1]*(int(not is_terminals[t])) - values[i]   # *(int(not is_terminals[t])) to eliminate the contribution in case state t is terminal
                    a_t += np.power(gamma*lambd, count) * delta_t
                    count += 1

                advantages[t] = a_t[0]

            batch_losses = []
            for batch in self.__generate_batches(memories, batch_size):
                with tf.GradientTape(persistent=True) as tape:
                    states_tf = tf.convert_to_tensor(states[batch])
                    actions_tf = tf.convert_to_tensor(actions[batch])
                    log_probs_tf = tf.convert_to_tensor(log_probs[batch])

                    probs = self.__actor(states_tf)
                    dist = tfp.distributions.Categorical(probs)
                    new_probs = dist.log_prob(actions_tf)

                    value = tf.squeeze(self.__critic(states_tf), 1)

                    prob_ratio = tf.math.exp(new_probs - log_probs_tf)
                    weighted_probs = advantages[batch]*prob_ratio

                    clipped_probs = tf.clip_by_value(prob_ratio, 1-epsilon_clip, 1+epsilon_clip)
                    weighted_clipped_probs = clipped_probs*advantages[batch]

                    actor_loss = tf.reduce_mean(-tf.math.minimum(weighted_probs, weighted_clipped_probs))
                    returns = advantages[batch] + values[batch]

                    critic_loss = tf.keras.losses.MSE(value, returns)

                gradients_actor = tape.gradient(actor_loss, self.__actor.trainable_variables)
                self.__actor.optimizer.apply_gradients(zip(gradients_actor, self.__actor.trainable_variables))

                gradients_critic = tape.gradient(critic_loss, self.__critic.trainable_variables)
                self.__critic.optimizer.apply_gradients(zip(gradients_critic, self.__critic.trainable_variables))

                batch_losses.append([np.mean(actor_loss), np.mean(critic_loss)])

            losses.append(np.mean(np.array(batch_losses), axis=0))

        return np.mean(np.array(losses), axis=0)
            


    
    def __generate_batches(self, memories, batch_size):
        states_len = len(memories["states"])
        batch_start = np.arange(0, states_len, batch_size)
        index = np.arange(states_len)
        np.random.shuffle(index)

        return [index[i:i+batch_size] for i in batch_start]
    
