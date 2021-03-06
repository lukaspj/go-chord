package chord

const fingerCount = 10

type fingerTable struct {
	fingers [fingerCount]*ContactInfo
	next    int
}

func (table *fingerTable) GetFinger(index int) *ContactInfo {
	return table.fingers[index]
}

func (table *fingerTable) SetFinger(index int, info *ContactInfo) bool {
	if table.fingers[index] == nil || !info.Id.Equals(table.fingers[index].Id) {
		logger.Debug("Setting finger %d to: %s", index, info.Id.String())
		table.fingers[index] = info
		return true
	}
	return false
}
