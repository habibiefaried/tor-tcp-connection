tcp-to-tor:
	cd examples && go build -ldflags="-linkmode external -extldflags -static" tcp-to-tor.go 
	cd examples && GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-linkmode external -extldflags -static" tcp-to-tor.go  