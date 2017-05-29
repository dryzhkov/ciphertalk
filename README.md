# ciphertalk

1. server:
    go run ciphertalk/main.go
2. client 1:
    go run ciphertalk/client/client.go --from=bar --to=foo --interval=2s
3. client 2:
    go run ciphertalk/client/client.go --from=foo --to=bar --listen-only=true
