package raft_main

import (
	"log"
	"raft/raft_request"
	state "raft/raft_state"
	statemachine "raft/state_machine"
)

// lock is already acquired
func ConvertToFollower(term int) {
	state.State.PersistentState.CurrentTerm = term
	state.State.PersistentState.VotedFor = -1
	state.WritePersState(state.State.PersistentState)
	state.State.Role = state.Follower

	log.Printf("Becomed follower at term %v\n", state.State.PersistentState.CurrentTerm)
}

// lock is already acquired
func Commit(until int) {
	until = min(until, len(state.State.PersistentState.Log))
	for i := state.State.VolatileState.LastApplied; i < until; i++ {
		entry := state.State.PersistentState.Log[i]
		is_err := false
		switch entry.Op {
		case "Create":
			is_err = statemachine.Create(entry.Uuid, entry.Body) == nil
		case "Update":
			is_err = statemachine.Update(entry.Uuid, entry.Body) == nil
		case "Delete":
			is_err = statemachine.Delete(entry.Uuid) == nil
		case "CAS":
			is_err = statemachine.CAS(entry.Uuid, entry.Body) == nil
		}
		state.State.PersistentState.Res = append(state.State.PersistentState.Res, is_err)
		log.Printf("commited index %v\n", i)
	}
	state.WritePersState(state.State.PersistentState)
	state.State.VolatileState.CommitIndex = max(until, state.State.VolatileState.CommitIndex)
	state.State.VolatileState.LastApplied = max(until, state.State.VolatileState.LastApplied)
}

func GetQuorumReadOnce(addr, key string, ch chan raft_request.RequestReadResp, exists bool) {
	for i := 0; i < 10; i++ {
		ret, err := raft_request.RequestRead(addr, key)
		if err == nil {
			ch <- ret
		}
	}
	ch <- raft_request.RequestReadResp{Value: "", Exists: !exists}
}

func GetQuorumRead(key string) (raft_request.RequestReadResp, bool) {
	ch := make(chan raft_request.RequestReadResp)

	value, err := statemachine.Read(key)
	me := raft_request.RequestReadResp{
		Value:  value,
		Exists: err == nil,
	}
	// log.Printf("%+v", me)

	for i := 0; i < state.ServerNum; i++ {
		if i == state.Me {
			continue
		}
		go GetQuorumReadOnce(state.Num2server_raft[i], key, ch, me.Exists)
	}

	ok := 1
	not_ok := 0
	for {
		b := <-ch
		if b.Exists == me.Exists && b.Value == me.Value {
			ok += 1
		} else {
			not_ok += 1
		}
		if ok*2 >= state.ServerNum {
			return me, true
		}
		if not_ok*2 >= state.ServerNum {
			return me, false
		}
	}
}
