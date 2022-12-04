
doc:
	godoc --http :8080

test:
	go test

new_version:
	./scripts/update_version.sh

config_git_hooks:
	git config core.hooksPath .githooks
