.PHONY: lint
# golangci-lint
lint:
	@for dir in $(shell find . -type f -name "*.mod" -print0 | xargs -0 realpath | xargs -n 1 dirname | sort -u); do \
	  echo $$dir processing...; \
		pushd $$dir > /dev/null && golangci-lint run --fix && popd > /dev/null; \
	done

.PHONY: tidy
# go mod tidy
tidy:
	@for dir in $(shell find . -type f -name "*.mod" -print0 | xargs -0 realpath | xargs -n 1 dirname | sort -u); do \
	  echo $$dir processing...; \
		pushd $$dir > /dev/null && go mod tidy && popd > /dev/null; \
	done

.PHONY: show
# show mod contains 'go-cinch'
show:
	@for dir in $(shell find . -type f -name "*.mod" -print0 | xargs -0 realpath | xargs -n 1 dirname | sort -u); do \
	  echo $$dir processing...; \
		pushd $$dir > /dev/null && cat go.mod | grep 'go-cinch' && popd > /dev/null; \
	done

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
