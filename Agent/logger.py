# main.py
import logging

def init(level = logging.INFO) -> logging.Logger:
    logging.basicConfig(
        level=level,
        format="{\"time\":\"%(asctime)s\",\"level\":\"%(levelname)s\",\"msg\":\"%(message)s\"",
        handlers=[
            logging.StreamHandler()
            #logging.FileHandler("app.log", encoding="utf-8") # File output
        ]
    )

    return logging.getLogger(__name__)

def init_debug():
    return init(level=logging.DEBUG)