package slave

import (
	"consts"
	"network"
	"log"
	"encoding/json"
	"sync"
	"time"
	"net"
)

type Backup struct {
	mux 		sync.Mutex
	backupData	consts.BackupSync
	conn 		*net.UDPConn
}

func (b *Backup) setBackupData(data consts.BackupSync)  {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.backupData = data
}

func (b *Backup) getBackupData() consts.BackupSync  {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.backupData
}

func (b *Backup) setMasterConn(conn *net.UDPConn)  {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.conn = conn
}

func (b *Backup) getMasterConn() *net.UDPConn  {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.conn
}

// listen for master data and store them as a backup
func (b *Backup) listenIncomingMsg() {
	var backupSync consts.BackupSync
	buffer := make([]byte, 8192)

	for {
		n, err := b.getMasterConn().Read(buffer[0:])
		if err != nil {
			log.Println(consts.Red, "master reading failed", consts.Neutral)
			return
		}
		//log.Println(consts.Cyan, buffer, consts.Neutral)
		if len(buffer) > 0 {
			//log.Println(consts.Cyan, string(buffer), consts.Neutral)
			err2 := json.Unmarshal(buffer[0:n], &backupSync)
			if err2 != nil {
				log.Println(consts.Red, "unmarshal backup failed", consts.Neutral)
				log.Fatal(err2)
			} else {
				b.setBackupData(backupSync)
			}
		}
	}
}

func (b *Backup) checkMasterAlive(masterFailed chan<- consts.BackupSync)  {
	for {
		data := b.getBackupData()
		if time.Since(data.Timestamp).Seconds() > 3 {
			log.Println(consts.Red, "Master died", consts.Neutral)
			b.getMasterConn().Close()
			masterFailed <- b.getBackupData()
			return
		}

		time.Sleep(10 * consts.PollRate)
	}
}

// Backup
/**
 * defer old instance
 * create Backup
 * listen for incoming DB syncs from Master
 * ping Master
 * -> became Master if prev Master failed
 * do same things as Slave
 */
func StartBackup(masterFailed chan<- consts.BackupSync) {
	conn := network.GetListenConn(network.GetBroadcastAddress()+consts.BackupPort)
	backup := Backup{}
	backup.setMasterConn(conn)

	go backup.listenIncomingMsg()

	// wait for master data before checking validity
	time.Sleep(200 * consts.PollRate)
	go backup.checkMasterAlive(masterFailed)
}
