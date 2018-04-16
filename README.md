# Elevator project for Team Bambuchistrom
- The elevator is programmed in Go

## Helpers
**Print dependency graph to help with the understanding:**
- `graphpkg -match 'common|master|slave|network|main|helper|consts'  elevator network main helper consts`

**Start elevator and override params:**
- `./SimElevatorServer --port 20001`
- `go build src/main/start.go; ./start

## Libraries used
-  We have only used the standard Go libraries
