package go_chord

const fingerCount = 10

// TODO better name
type ContactInfo struct {
	Address string
	Id      NodeID
}

type fingerTable struct {
	fingers       [fingerCount]ContactInfo
	next          int
}

/*
func (table *fingerTable) GetSuccessor() ContactInfo {
	return table.fingers[0]
}

func (table *fingerTable) SetSuccessor(info ContactInfo) {
	if !info.Id.Equals(table.fingers[0].Id) {
		logger.Info("Setting successor to: %s", info.Id.String())
		table.fingers[0] = info
		table.lastDirtyTime = time.Now()
	}
}
*/

func (table *fingerTable) GetFinger(index int) ContactInfo {
	return table.fingers[index]
}

func (table *fingerTable) SetFinger(index int, info ContactInfo) bool {
	if !info.Id.Equals(table.fingers[index].Id) {
		logger.Info("Setting finger %d to: %s", index, info.Id.String())
		table.fingers[index] = info
		return true
	}
	return false
}
