from typing import Any
from itertools import chain
from awslambdaric.lambda_context import LambdaContext as LambdaContextType
from utils.connection import ConnectionHolder
from utils.logger import init_logger
from utils.cipher import get_blob, decrypt_data_key, encrypt
import pyarrow.parquet as pq
import pymysql.cursors
import json
import tomllib


CONFIG_PATH = "./config.toml"
REQUIRE_EVENT_PAYLOAD = [
    "Product",
    "Customer",
    "Store",
]


def validate(event: dict[str, str]) -> None:
    for payload_field in REQUIRE_EVENT_PAYLOAD:
        assert payload_field in event, f"Bad request: not enough event payload '{payload_field}'"


def lambda_handler(event: dict[str, str], context: LambdaContextType) -> str:
    logger = init_logger(context.function_name)
    logger.info(json.dumps(event))

    validate(event)

    with open(CONFIG_PATH, "rb") as f:
        config = tomllib.load(f)
        param_store_name = config["param_store_name"]
        connection_name = config["connection_name"]
        del config["param_store_name"]
        del config["connection_name"]

    cipher_text_blob = get_blob(param_store_name, with_decryption=True)
    data_key = decrypt_data_key(cipher_text_blob)

    def migrate_rows(database: str, table: str, rows: list[dict[str, Any]], cursor: pymysql.cursors.DictCursor) -> dict[str, Any]:
        logger.info(f"migration to '{database}.{table}'")

        has_enc = "encryption_columns" in config[database][table]
        encryption_cols: list[str] = config[database][table]["encryption_columns"] if has_enc else []
        migration_cols: list[str] = config[database][table]["migration_columns"]

        lcols = [col.lower() for col in migration_cols]
        column_stmt = ", ".join(lcols)
        values_stmt = ", ".join([f"({', '.join(['%s' for _ in migration_cols])})" for _ in rows])
        update_stmt = ", ".join([f"{col} = VALUES({col})" for col in lcols if col != "id"])

        values = [[encrypt(row[col], data_key) if col in encryption_cols else row[col] for col in migration_cols] for row in rows]
        values = tuple(chain.from_iterable(values))

        upsert_query = f"INSERT INTO {table} ({column_stmt}) VALUES {values_stmt} ON DUPLICATE KEY UPDATE {update_stmt}"
        cursor.execute(upsert_query, values)

    connection_meta = ConnectionHolder(connection_name)
    conn = pymysql.connect(host=connection_meta.hostname,
                           user=connection_meta.username,
                           password=connection_meta.password,
                           database="confidential",
                           cursorclass=pymysql.cursors.DictCursor)
    with conn:
        try:
            with conn.cursor() as cursor:
                for key, subconf in config.items():
                    arrow_table = pq.read_table(event[key])
                    migrate_rows(subconf["database"], subconf["table"], arrow_table.to_pylist(), cursor)
            conn.commit()
        except Exception:
            conn.rollback()
            raise

    return "Function Succeeded."
