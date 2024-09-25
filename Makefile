HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

gen: 
	buf generate --path ./api -o ${HOME}

go-vendor:
	go mod tidy && go mod vendor

build: clean
	# env GOOS=linux GOARCH=arm go build -mod vendor -o $(BIN)
	env GOOS=linux GOARCH=amd64 go build -mod vendor -o $(BIN)

docker-build:
	docker build --no-cache --progress=plain -t trading:1.0.0 -f Dockerfile .

docker-compose:
	docker-compose up --detach --build

run-local:
	go run main.go start --config config.dev.toml --env .env.local

rsync:
	rsync -avz futures-trading *.toml .env runserver* admin@18.143.73.197:/server/futures-trading

dockerhub:
	docker tag anvh2/futures-trading:v1.0.1 anvh2/futures-trading:v1.0.1
	docker push anvh2/futures-trading:v1.0.1