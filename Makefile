.PHONY: test bench

test: ## Runs all unit tests
	go test -v -race -count='1' ./...

bench: ## Runs all benchmarks
	go test -v -run='^a' -bench='.' -benchtime='50x' -benchmem ./...
