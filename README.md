# Elevator project for Team Bambuchistrom

## Helpers
**Print dependency graph:**
- `graphpkg -match 'common|master|slave|network|main|helper|consts'  elevator network main helper consts`

**Start elevator and override params:**
- `./SimElevatorServer --port 20001`
- `go build src/main/start.go; ./start -masterPort 40000 -myPort 20000 -elPort 20001 -myRole 1`


## TODO list
### Fault tolerance 
- [ ] Elevator fails to communicate with controller
- [ ] Slave fails to communicate with Master
- [ ] Backup fails to communicate with Master
- [ ] Master fails to communicate with Slave
- [ ] Master fails to communicate with Backup

### Network stuff
- [ ] Role decision (who is Master/Backup/Slave)
- [ ] Master
  - [x] Broadcast its IP address once Master is elected
  - [x] Receive Slave's IP addresses and store them
    - Create UDPConn from IP address as well
  - [ ] Send periodic info about all elevators to Backup
- [ ] Backup
  - [ ] Ping Master 
  - [ ] Became Master when previous is down
  - [ ] Receive global knowledge and store it
- [ ] Common (for all roles)
  - [x] Receive Master IP address and store it as UDPConn
  - [x] Notify Master and send him own IP address
