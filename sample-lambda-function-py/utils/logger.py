import sys
import logging
from datetime import timezone, timedelta, datetime


class ISOTimeFormatter(logging.Formatter):

    def formatTime(self, record: logging.LogRecord, datefmt=None):
        tz_jst = timezone(timedelta(hours=+9), "JST")
        ct = datetime.fromtimestamp(record.created, tz=tz_jst)
        s = ct.isoformat(timespec="microseconds")
        return s


def init_logger(symbol: str) -> logging.Logger:
    logger = logging.getLogger()
    [logger.removeHandler(h) for h in logger.handlers]
    formatter = ISOTimeFormatter(
        f"[%(asctime)s][%(levelname)s][{symbol}:%(funcName)s:%(lineno)s] %(message)s\n",
        "%Y-%m-%d %H:%M:%S",
    )
    stdout_handler = logging.StreamHandler(stream=sys.stdout)
    stdout_handler.setFormatter(formatter)
    logger.addHandler(stdout_handler)
    logger.setLevel(logging.INFO)
    logger.propagate = False
    return logger