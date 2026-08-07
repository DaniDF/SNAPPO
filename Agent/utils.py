import os

from actor import ActorCritic

MODEL_WEIGHTS_CHECKPOINT_PATH_KEY = "model_weights_checkpoint_path"
MODEL_CHECKPOINT_PATH_KEY = "model_checkpoint_path"
CHECKPOINT_INFO_FILENAME = "checkpoint"

MODEL_FILENAME = "model.keras"

def save_policy(policy: ActorCritic, dir: str, only_weights = True):
    dir_paths = {
        "actor": os.path.join(dir,"actor"),
        "critic": os.path.join(dir,"critic")
    }

    for type, path in dir_paths.items():
        os.makedirs(path, exist_ok=True)

        new_filename = __get_next_ckpt_weights_name(path)
        data_update = {MODEL_WEIGHTS_CHECKPOINT_PATH_KEY: new_filename}
        if type == "actor":
            policy.save_actor_model_weights(os.path.join(path,new_filename))

            if not only_weights:
                policy.save_actor_model(os.path.join(path,MODEL_FILENAME))
                data_update[MODEL_CHECKPOINT_PATH_KEY] = MODEL_FILENAME
        else:
            policy.save_critic_model_weights(os.path.join(path,new_filename))
            if not only_weights:
                policy.save_critic_model(os.path.join(path,MODEL_FILENAME))
                data_update[MODEL_CHECKPOINT_PATH_KEY] = MODEL_FILENAME


        __update_checkpoint_file(os.path.join(path,CHECKPOINT_INFO_FILENAME), new_values=data_update)

def load_policy(policy: ActorCritic, dir: str, only_weights=True):
    dir_paths = {
        "actor": os.path.join(dir,"actor"),
        "critic": os.path.join(dir,"critic")
    }

    for type, path in dir_paths.items():
        ckpt_data = __read_checkpoint_file(os.path.join(path,CHECKPOINT_INFO_FILENAME))

        if type == "actor":
            if not only_weights:
                policy.load_actor_model(os.path.join(path, ckpt_data[MODEL_CHECKPOINT_PATH_KEY]))

            policy.load_actor_model_weights(os.path.join(path, ckpt_data[MODEL_WEIGHTS_CHECKPOINT_PATH_KEY]))
        else:
            if not only_weights:
                policy.load_critic_model(os.path.join(path, ckpt_data[MODEL_CHECKPOINT_PATH_KEY]))

            policy.load_critic_model_weights(os.path.join(path, ckpt_data[MODEL_WEIGHTS_CHECKPOINT_PATH_KEY]))


def __get_last_ckpt_weights_name(ckpt_path: str = ".") -> str:
    try:
        checkpoint_data = __read_checkpoint_file(os.path.join(ckpt_path, "checkpoint"))
        result = checkpoint_data[MODEL_WEIGHTS_CHECKPOINT_PATH_KEY]
        
    except FileNotFoundError or KeyError:
        result = None

    return result


def __get_next_ckpt_weights_name(ckpt_path: str = ".") -> str:
    ckpt_last_file = __get_last_ckpt_weights_name(ckpt_path)

    if ckpt_last_file is not None:
        ckpt_last_num = int(ckpt_last_file[3:6])
        new_filename = f"cp-{ckpt_last_num + 1:03d}.ckpt.weights.h5"
    else:
        new_filename = "cp-000.ckpt.weights.h5"

    return new_filename

def __read_checkpoint_file(path) -> dict[str,str]:
    result = {}

    with open(path, "r") as file_in:
        for line in file_in.readlines():
            split = line.split(":")
            key = split[0].strip()
            value = split[1].strip().replace("\"", "").strip()

            result[key] = value

    return result

def __update_checkpoint_file(path, new_values: dict[str,str]):
    try:
        ckpt_data = __read_checkpoint_file(path)

    except FileNotFoundError:
        ckpt_data = {}

    for key, value in new_values.items():
        ckpt_data[key] = value

    with open(path, "w") as file_out:
        for key, value in ckpt_data.items():
            file_out.write(f"{key}: \"{value}\"\n")