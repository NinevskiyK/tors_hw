package raft_main

import (
	"log"
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
