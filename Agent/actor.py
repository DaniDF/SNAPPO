import tensorflow as tf
import tensorflow_probability as tfp
import numpy as np

class ActorCritic(tf.keras.Model):
    def __init__(self, input_shape: tuple[int], action_dim: int, learning_rate = 3e-4):
        super(ActorCritic, self).__init__()
        
        # Defining the actor model
        self.__actor = tf.keras.models.Sequential([
            tf.keras.layers.Input(shape=input_shape, dtype=tf.float32),
            tf.keras.layers.Dense(512, activation="relu", dtype=tf.float32),  # TODO Evaluate swish or gelu
            tf.keras.layers.Dense(256, activation="relu", dtype=tf.float32),
            tf.keras.layers.Dense(128, activation="relu", dtype=tf.float32),
            tf.keras.layers.Dense(action_dim, activation="softmax")
        ])
        self.__actor.compile(optimizer=tf.keras.optimizers.Adam(learning_rate))
        
        # Defining the critic model
        self.__critic = tf.keras.models.Sequential([
            tf.keras.layers.Input(shape=input_shape, dtype=tf.float32),
            tf.keras.layers.Dense(512, activation="relu", dtype=tf.float32),  # TODO Evaluate swish or gelu
            tf.keras.layers.Dense(256, activation="relu", dtype=tf.float32),
            tf.keras.layers.Dense(128, activation="relu", dtype=tf.float32),
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
        value = self.__critic(tf.convert_to_tensor([state]), dtype=tf.float32)
        return tf.squeeze(value).numpy().item()


    def ppo_update(self, memories, batch_size=10, gamma=0.99, lambd=0.95, epsilon_clip=0.2, k_epochs=10, entropy_coef=0.01):
        states = np.array(memories["states"], dtype=np.float32)
        actions = np.array(memories["actions"], dtype=np.float32)
        values = np.array(memories["values"], dtype=np.float32)
        log_probs = np.array(memories["log_probs"], dtype=np.float32)
        rewards = np.array(memories["rewards"], dtype=np.float32)
        is_terminals = np.array(memories["is_terminals"], dtype=np.float32)

        advantages = np.zeros(len(rewards), dtype=np.float32)
        last_gae_lam = 0
        
        for t in reversed(range(len(rewards))):
            next_non_terminal = 1.0 - is_terminals[t]
            if t == len(rewards) - 1:
                next_values = 0.0
            else:
                next_values = values[t + 1]
                
            delta = rewards[t] + gamma * next_values * next_non_terminal - values[t]
            last_gae_lam = delta + gamma * lambd * next_non_terminal * last_gae_lam
            advantages[t] = last_gae_lam

        returns = advantages + values

        advantages = (advantages - np.mean(advantages)) / (np.std(advantages) + 1e-8) # Normalization, 1e-8 ensure not to divide by 0

        losses = []

        for _ in range(k_epochs):
            batch_losses = []
            
            for batch in self.__generate_batches(memories, batch_size):
                with tf.GradientTape(persistent=True) as tape:
                    states_tf = tf.convert_to_tensor(states[batch])
                    actions_tf = tf.convert_to_tensor(actions[batch])
                    log_probs_tf = tf.convert_to_tensor(log_probs[batch])
                    returns_tf = tf.convert_to_tensor(returns[batch])
                    advantages_tf = tf.convert_to_tensor(advantages[batch])

                    probs = self.__actor(states_tf)
                    dist = tfp.distributions.Categorical(probs)
                    new_probs = dist.log_prob(actions_tf)
                    entropy = dist.entropy()

                    value = tf.squeeze(self.__critic(states_tf), 1)

                    prob_ratio = tf.math.exp(new_probs - log_probs_tf)
                    
                    weighted_probs = advantages_tf * prob_ratio
                    clipped_probs = tf.clip_by_value(prob_ratio, 1 - epsilon_clip, 1 + epsilon_clip)
                    weighted_clipped_probs = clipped_probs * advantages_tf
                    
                    actor_loss = tf.reduce_mean(-tf.math.minimum(weighted_probs, weighted_clipped_probs))
                    actor_loss -= entropy_coef * tf.reduce_mean(entropy)

                    critic_loss = tf.reduce_mean(tf.square(returns_tf - value))

                gradients_actor = tape.gradient(actor_loss, self.__actor.trainable_variables)
                self.__actor.optimizer.apply_gradients(zip(gradients_actor, self.__actor.trainable_variables))

                gradients_critic = tape.gradient(critic_loss, self.__critic.trainable_variables)
                self.__critic.optimizer.apply_gradients(zip(gradients_critic, self.__critic.trainable_variables))
                
                del tape 

                batch_losses.append([np.mean(actor_loss), np.mean(critic_loss)])

            losses.append(np.mean(np.array(batch_losses), axis=0))

        return np.mean(np.array(losses), axis=0)

    
    def __generate_batches(self, memories, batch_size):
        states_len = len(memories["states"])
        batch_start = np.arange(0, states_len, batch_size)
        index = np.arange(states_len)
        np.random.shuffle(index)

        return [index[i:i+batch_size] for i in batch_start]
    

    def save_actor_model_weights(self, filename):
        self.__actor.save_weights(filename)

    def save_critic_model_weights(self, filename):
        self.__critic.save_weights(filename)
    
    def load_actor_model_weights(self, filename):
        self.__actor.load_weights(filename)

    def load_critic_model_weights(self, filename):
        self.__critic.load_weights(filename)

    def save_actor_model(self, filename):
        self.__actor.save(filename)

    def save_critic_model(self, filename):
        self.__critic.save(filename)

    def load_actor_model(self, filename):
        self.__actor = tf.keras.models.load_model(filename)

    def load_critic_model(self, filename):
        self.__critic = tf.keras.models.load_model(filename)