param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ComposeArgs
)

docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build @ComposeArgs