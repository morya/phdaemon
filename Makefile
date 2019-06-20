.phony:all

all:phdaemon

phdaemon: $(wildcard *.go)
	@gofmt -w $<
	@go build 

clean:
	@rm -rf phdaemon
