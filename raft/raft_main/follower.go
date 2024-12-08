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
		switch entry.Op {
		case "Create":
			statemachine.Create(entry.Uuid, entry.Body)
		case "Update":
			statemachine.Update(entry.Uuid, entry.Body)
		case "Delete":
			statemachine.Delete(entry.Uuid)
		}
		log.Printf("commited index %v\n", i)
	}
	state.State.VolatileState.CommitIndex = max(until, state.State.VolatileState.CommitIndex)
	state.State.VolatileState.LastApplied = max(until, state.State.VolatileState.LastApplied)
}
