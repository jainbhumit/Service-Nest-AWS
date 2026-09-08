.PHONY: build-HelloWorldFunction

# SAM invokes this target when Metadata.BuildMethod is makefile.
# provided.al2023 requires the executable to be named bootstrap.
build-HelloWorldFunction:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o "$(ARTIFACTS_DIR)/bootstrap" ./cmd
