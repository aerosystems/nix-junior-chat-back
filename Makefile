## up: starts all containers in the background without forcing build
up:
	@echo "Starting docker images..."
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev up -d
	@echo "Docker images started!"

## down: stop docker compose
down:
	@echo "Stopping docker images..."
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev down
	@echo "Docker stopped!"

## rebuild: rebuilding all containers without cache
rebuild:
	@echo "Rebuilding docker images..."
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev down
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev build --no-cache
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev up -d
	@echo "Docker images rebuilt!"

## chat: stops chat-service, removes docker image, builds service, and starts it
chat: build
	@echo "Building chat-service docker image..."
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev stop chat-service
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev rm -f chat-service
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev up --build -d chat-service
	docker-compose -f ./docker-compose.dev.yml --env-file ./.env.dev start chat-service
	@echo "chat-service built and started!"

## build: builds the chat-service binary as a linux executable
build:
	@echo "Building chat-service binary.."
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o chat-service.bin ./cmd/app
	@echo "App chat-service was built!"

## start: start application
start:
	go run ./cmd/app/*
	
## doc: generating Swagger Docs
doc:
	@echo "Stopping generating Swagger Docs..."
	swag init -g ./cmd/app/* --output ./docs
	@echo "Swagger Docs prepared, look at /docs"

## help: displays help
help: Makefile
	@echo " Choose a command:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'