run.%:
	@go run cmd/$*/main.go

build.docker:
	@docker build -f build/package/healthcheck.Dockerfile -t healthcheck .


