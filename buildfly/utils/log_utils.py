import logging
import os

BLACK, RED, GREEN, YELLOW, BLUE, MAGENTA, CYAN, WHITE = range(8)
RESET_SEQ = "\033[0m"
COLOR_SEQ = "\033[1;%dm"
BOLD_SEQ = "\033[1m"

COLORS = {
    'WARNING': YELLOW,
    'INFO': WHITE,
    'DEBUG': BLUE,
    'CRITICAL': YELLOW,
    'ERROR': RED
}


class ColoredFormatter(logging.Formatter):
    # https://docs.python.org/3/library/logging.html#logrecord-attributes
    FORMAT = "[%(asctime)s][%(levelname)s] %(message)s "
    DBG_FORMAT = "[%(asctime)s][%(levelname)s][%(name)s][%(filename)s:%(lineno)d] %(message)s "


    def __init__(self, use_color=True):
        dbg = os.environ.get("BFLY_DEBUG", "false").lower() == "true"
        logging.Formatter.__init__(self, self.DBG_FORMAT if dbg else self.FORMAT)
        self.use_color = use_color

    def format(self, record):
        levelname = record.levelname
        if self.use_color and levelname in COLORS:
            levelname_color = COLOR_SEQ % (30 + COLORS[levelname])
            record.msg = levelname_color + record.msg + RESET_SEQ
        return logging.Formatter.format(self, record)


class BFLogger(logging.Logger):
    def __init__(self, name):
        logging.Logger.__init__(self, name, os.environ.get("BFLY_LOGGER_LEVEL", "INFO"))
        use_color = True
        if os.environ.get("USE_COLOR_LOGGER", "true").lower() == "false":
            use_color = False
        color_formatter = ColoredFormatter(use_color)
        console = logging.StreamHandler()
        console.setFormatter(color_formatter)
        self.addHandler(console)

logging.setLoggerClass(BFLogger)

def get_logger(name):
    return logging.getLogger(name)
