# Contributing to pod-reaper

## Issues

Issues are always welcome! You can expect conversation.

## Creating a rule

- create tests for the rule
- implement the rule interface in ./rules/rules.go
- add the rule `LoadRules` method in ./rules/rules.go
- ensure your tests pass `go test ./rules`
- format using `go fmt ./rules`
- run the linter `golint` (https://github.com/golang/lint)

## Pull Requests

These rules must be followed for any contributions to be merged into master. A Git installation is required.
See [here](./docs/getting_started_git.md) for more information.

1. Fork this repo
1. Create a branch
1. Make an desired changes
1. Validate you changes meet your desired use case
1. Ensure documentation has been updated
1. Format you changes `go fmt ./reaper ./rules`
1. Run a go linter with `golint` (https://github.com/golang/lint)
1. Open a pull-request: you can expect discussion
