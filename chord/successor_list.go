package chord

const successorListSize = 5

type successorList [successorListSize]ContactInfo

func (successors *successorList) GetSuccessor(i int) ContactInfo {
	return successors[i]
}

func (successors *successorList) SetSuccessor(i int, info ContactInfo) bool {
	if !info.Id.Equals(successors[i].Id) {
		logger.Info("Setting successor %d to: %s", i, info.Id.String())
		successors[i] = info
		return true
	}
	return false
}