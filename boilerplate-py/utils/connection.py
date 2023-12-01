import boto3
from urllib.parse import urlparse


class ConnectionHolder:
    connection: dict

    def __init__(self, connection_name: str):
        self.connection = boto3.client("glue").get_connection(Name=connection_name)

    @property
    def connection_name(self) -> str:
        return self.connection["Connection"]["Name"]

    @property
    def username(self) -> str:
        return self.connection["Connection"]["ConnectionProperties"]["USERNAME"]

    @property
    def password(self) -> str:
        return self.connection["Connection"]["ConnectionProperties"]["PASSWORD"]

    @property
    def hostname(self) -> str:
        url: str = self.connection["Connection"]["ConnectionProperties"]["JDBC_CONNECTION_URL"]
        if url.startswith("jdbc:"):
            url = url.replace("jdbc:", "")
        parsed_url = urlparse(url)
        return parsed_url.hostname

    def url_with_db(self, database: str) -> str:
        url = self.connection["Connection"]["ConnectionProperties"]["JDBC_CONNECTION_URL"].split("/")
        url[-1] = database
        return "/".join(url)
