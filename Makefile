# ==================================================================================== #
# BUILD
# ==================================================================================== #

run:
	@go run . --config "config"

test:
	@go test -v

prod:
	@GOOS=linux GOARCH=amd64 go build -o pr-discord-notifier 
	@rsync -avP pr-discord-notifier config.toml tinkaling:~/pr-discord-notifier
	@rsync -avP pr-discord-notifier.service Caddyfile tinkaling:~
	ssh -t tinkaling "\
		 sudo mv ~/pr-discord-notifier.service /etc/systemd/system/ && \
		 sudo systemctl enable pr-discord-notifier && \
		 sudo systemctl restart pr-discord-notifier && \
		 sudo mv ~/Caddyfile /etc/caddy/ && \
		 sudo systemctl reload caddy \
	"
	@rm pr-discord-notifier
