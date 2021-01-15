#test:
#	go test -v $(shell go list ./... | grep -v /vendor/) 
#
#testacc:
#	TF_ACC=1 go test -v ./wug -run="TestAcc"

build: deps
	gox -osarch="linux/amd64 windows/amd64 darwin/amd64 freebsd/amd64" \
	-output="pkg/{{.OS}}_{{.Arch}}/terraform-provider-wug" .

deps:
#	go get -u github.com/hashicorp/terraform/plugin
	
clean:
	rm -rf pkg/

default: build
