b:
	flags="-X 'version.GoVersion=$(go version)'"
	go build -ldflags "$flags" -x -o main -version main.go

enc:
	./main ss --enc -k 123456 -t hello

dec:
	./main ss --dec -k 123456 -t WVdZWFo=

r:
	./main s
