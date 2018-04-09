# Elevator project for Team Bambuchistrom

## TODO list
### Network stuff
- [ ] Role decision (who is Master/Backup/Slave)
- [ ] Master
  - [ ] Broadcast its IP address once Master is elected
  - [ ] Receive Slave's IP addresses and store them
    - Create UDPConn from IP address as well
  - [ ] Send periodic info about all elevators to Backup
- [ ] Backup
  - [ ] Ping Master 
  - [ ] Became Master when previous is down
  - [ ] Receive global knowledge and store it
- [ ] Common (for all roles)
  - [ ] Receive Master IP address and store it as UDPConn
  - [ ] Notify Master and send him own IP address
