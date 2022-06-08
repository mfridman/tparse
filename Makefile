preview:
	mkdocs serve

build:
	rm -rf site ./.cache \
	&& mkdocs build --config-file mkdocs.yml
