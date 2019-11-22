b:
	go build main.go

enc:
	./main ss --enc -k 123456 -t hello

dec:
	./main ss --dec -k 123456 -t WVdZWFo=

r:
	./main s
