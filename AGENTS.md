# Instructions to Agents for Go Wordle App

## Project Structure

The desired structure is that there should be a `./pkg/` directory, with a package called `wordlegameengine`. There should be a `main.go` file in the project root. Other packages or folders may be added as necessary, but it is likely we will only need one package. Business logic does not go in `main.go` - business logic goes in the `wordlegameengine` package.

## Language version

The language version is Go 1.26. You MAY NOT target other language versions, or attempt to change this.

## Running Unit Tests

If you are supposed to run the tests, use `go test ./...` .

## Data

The data files are the two text files in `./data/`. You MUST NOT change these.
