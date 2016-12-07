all:
	-mkdir -p bin
	cd kr; go build -o ../bin/kr
	cd krd; go build -o ../bin/krd
	cd pkcs11; make; cp kr-pkcs11.so ../bin/kr-pkcs11.so

check:
	go test github.com/agrinman/kr github.com/agrinman/kr/pkcs11 github.com/agrinman/kr/krd github.com/agrinman/kr/krdclient github.com/agrinman/kr/kr
