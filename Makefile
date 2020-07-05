.PHONY: lrparser lr1parser

lrparser:
	go run ./cmd/lrparser/main.go

lr1parser:
	go run ./cmd/lr1parser/main.go
