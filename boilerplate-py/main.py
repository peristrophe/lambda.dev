from awslambdaric.lambda_context import LambdaContext as LambdaContextType


REQUIRE_EVENT_PAYLOAD = [
    "Hoge",
    "Fuga",
    "Piyo",
]


def validate(event: dict[str, str]) -> None:
    for payload_field in REQUIRE_EVENT_PAYLOAD:
        assert payload_field in event, f"Bad request: not enough event payload '{payload_field}'"


def lambda_handler(event: dict[str, str], context: LambdaContextType) -> str:
    validate(event)
    return "Function Succeeded."
