default:: ui build

clean::
	rm -rf dist/

ui::
	npm run build

build::
	go build

snapshot::
	goreleaser --snapshot --clean
