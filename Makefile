run.%:
	@go run cmd/$*/main.go

build:
	@docker build -f build/package/Dockerfile -t dafaque_job --build-arg APP=job .
	@docker build -f build/package/Dockerfile -t dafaque_api --build-arg APP=api .

up:
	@docker-compose -f build/ci/docker-compose.yml up

