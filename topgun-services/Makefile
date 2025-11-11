all: go-build \
	docker-build \
	docker-save \
	docker-clean

go-build:
	CGO_ENABLED=0 go build -v \
	-buildvcs=false \
	-installsuffix 'static' \
	-ldflags="-X 'main.version=$$(git rev-parse --short HEAD)' -X 'main.build=$$(date --iso-8601=seconds)'" \
	-o ./dist/server \
	./cmd/server

docker-build:
	docker build -f ./cmd/server/Dockerfile \
	-t topgun-services \
	--pull \
	.

docker-save:
	docker save topgun-services | gzip > dist/topgun-services.tar.gz

docker-clean:
	docker image prune -f

go-mod:
	go get -v -u && go mod tidy

go-test:
	go test -v ./...

swag:
	swag init -g ./cmd/server/main.go 

go-run:
	go run ./cmd/server/

delete-build:
	rm -rf ./dist/server

air:
	air --build.cmd "make delete-build && make go-build" --build.bin "./dist/server"

