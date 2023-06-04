# Desktop App

This app demonstrates the GUI supported pairing process using the [eebus-go library](https://github.com/enbility/eebus-go). It consists of a server component written in Go and a web implemented using VueJS 3.

Once this app is paired with another service (could also be the same service running on a different port), it will show the SPINE data with details about supported usecases and features. The goal is to also present this information in a more user friendly way in the future.

Another goal is to provide an executable for every supported platform that contains everything required.

The service requires a certificate and a key which will be created automatically and saved in the working folder if file names are not provided or the default filenames are not found.

## First steps

- Download and install [golang](https://go.dev) for your computer
- Download and install [NodeJS and NPM](https://nodejs.org/) if you do not already have it
- Download the source code of this repository
- Run `npm install` inside the root repository folder
- Now follow either the `Development` or `Build binary` steps

## Development

### Running the server component

- `go run main.go -h` to see all the possible parameters.
- `go run main.go` to start with the default parameters.

### Running the web frontend

`npx vite dev` to start with the default parameters using `vite.config.js`. The web service is now accessible at `http://localhost:7051/`

## Build binary

- `make ui` for creating the UI assets
- `make build` for building the binary for the local system
- execute the binary with `./desktop-app`
- Open the website in a browser at `http://localhost:7050/`
