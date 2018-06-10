package chord

type ChordApi struct {
	peer *Peer
}

type RPCHeader struct {
	Sender     ContactInfo
	ReceiverId NodeID
}

type PingRequest struct {
	RPCHeader
}

type PingResponse struct {
	RPCHeader
}

func (pw *ChordApi) Ping(args *PingRequest, response *PingResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		logger.Info("Ping from %s\n", args.RPCHeader)
	} else {
		logger.Error("error happened: %v\n", err)
	}
	return
}

type FindSuccessorRequest struct {
	RPCHeader
	Id NodeID
}

type FindSuccessorResponse struct {
	RPCHeader
	Info ContactInfo
}

func (pw *ChordApi) FindSuccessor(args FindSuccessorRequest, response *FindSuccessorResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		var info ContactInfo
		info, err = pw.peer.FindSuccessor(args.Id)
		response.Info = info
	} else {
		logger.Error("error happened: %v", err)
	}
	return
}

type ClosestPrecedingNodeRequest struct {
	RPCHeader
	Id NodeID
}

type ClosestPrecedingNodeResponse struct {
	RPCHeader
	Info ContactInfo
}

func (pw *ChordApi) ClosestPrecedingNode(args *ClosestPrecedingNodeRequest, response *ClosestPrecedingNodeResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {

		for i := fingerCount - 1; i >= 0; i-- {
			finger := pw.peer.network.fingerTable.fingers[i]
			if finger.Id.Between(pw.peer.Info.Id, args.Id) {
				response.Info = finger
				return
			}
		}

		response.Info = pw.peer.Info
	} else {
		logger.Error("error happened: %v", err)
	}
	return
}

type PredecessorRequest struct {
	RPCHeader
}

type PredecessorResponse struct {
	RPCHeader
	Info *ContactInfo
}

func (pw *ChordApi) Predecessor(args *PredecessorRequest, response *PredecessorResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		response.Info = pw.peer.network.predecessor
	} else {
		logger.Error("error happened: %v", err)
	}
	return
}

type SuccessorRequest struct {
	RPCHeader
}

type SuccessorResponse struct {
	RPCHeader
	Info ContactInfo
}

func (pw *ChordApi) Successor(args *SuccessorRequest, response *SuccessorResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		response.Info = pw.peer.network.successors.GetSuccessor(0)
	} else {
		logger.Error("error happened: %v", err)
	}
	return
}

type NotifyRequest struct {
	RPCHeader
}

type NotifyResponse struct {
	RPCHeader
}

func (pw *ChordApi) Notify(args *NotifyRequest, response *NotifyResponse) (err error) {
	if err = pw.peer.HandleRPC(&args.RPCHeader, &response.RPCHeader); err == nil {
		if pw.peer.network.predecessor == nil || args.Sender.Id.Between(pw.peer.network.predecessor.Id, args.Sender.Id) {
			pw.peer.network.predecessor = &args.Sender
			pw.peer.Poke()
		}
	} else {
		logger.Error("error happened: %v", err)
	}
	return
}
