gen:
	go build -o bili-suit-tool main.go

amd64linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bili-suit-tool main.go
	tar zcvf bili-suit-tool-linux-amd64.tar.gz bili-suit-tool config.json README.md

arm64linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bili-suit-tool main.go
	tar zcvf bili-suit-tool-linux-arm64.tar.gz bili-suit-tool config.json README.md

amd64windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bili-suit-tool.exe main.go
	tar zcvf bili-suit-tool-windows-amd64.tar.gz bili-suit-tool.exe config.json README.md

amd64mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bili-suit-tool main.go
	tar zcvf bili-suit-tool-macOS-amd64.tar.gz bili-suit-tool config.json README.md

clean:
	rm bili-suit-tool*

