# ciphertalk

1. server:
    go run ciphertalk/main.go
2. client 1:
    go run ciphertalk/client/client.go --from=bar --to=foo --interval=2s
3. client 2:
    go run ciphertalk/client/client.go --from=foo --to=bar --listen-only=true


## Testing

1. run tests:
    go test ciphertalk/server/auth
2. run test and generate output file
    go test ciphertalk/server/auth -coverprofile=auth_cover.out
3. analyse test coverage
    go tool cover -html=auth_cover.out
