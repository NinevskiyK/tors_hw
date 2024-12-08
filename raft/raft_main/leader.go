package raft_main

import (
	"errors"
	"log"
	"raft/raft_request"
	state "raft/raft_state"
	"time"
)

func AddToLog(entry state.Log) bool {
	state.State.Mu.Lock()
	if state.State.Role != state.Leader {
		state.State.Mu.Unlock()
		return false
	}
	entry.Term = state.State.PersistentState.CurrentTerm

	state.State.PersistentState.Log = append(state.State.PersistentState.Log, entry)
	state.WritePersState(state.State.PersistentState)
	state.State.Mu.Unlock()
	return ReplicateLog(8)
}

// lock is already acquired
func AdvanceCommit(index int) {
	rep_num := 1 // leader
	for i := 0; i < state.ServerNum; i++ {
		if i == state.Me {
			continue
		}
		if state.State.VolatileLeaderState.MatchIndex[i] >= index {
			rep_num += 1
		}
	}
	if rep_num*2 > state.ServerNum {
		if index > state.State.VolatileState.CommitIndex {
			state.State.VolatileState.CommitIndex = index
			Commit(index)
			log.Printf("advance commit: %v", index)
		}
	}
}

func ReplicateLogOne(index int, ch chan bool, times int) {
	for i := 0; i < times; i++ {
		resp, match_index, err := raft_request.AppendEntries(index, state.Num2server_raft[index])
		if err != nil {
			ConvertToFollower(state.State.PersistentState.CurrentTerm)
			ch <- false
			return // not a leader
		}
		if resp.Term == -1 {
			continue // deadline
		}

		state.State.Mu.Lock()
		if resp.Term > state.State.PersistentState.CurrentTerm {
			ConvertToFollower(state.State.PersistentState.CurrentTerm)
			state.State.Mu.Unlock()
			ch <- false
			return
		}
		if resp.Success {
			state.State.VolatileLeaderState.MatchIndex[index] = max(match_index, state.State.VolatileLeaderState.MatchIndex[index])
			state.State.VolatileLeaderState.NextIndex[index] = max(match_index, state.State.VolatileLeaderState.NextIndex[index])
			AdvanceCommit(state.State.VolatileLeaderState.MatchIndex[index])
			state.State.Mu.Unlock()
			ch <- true
			return
		}
		times += 1
		state.State.VolatileLeaderState.NextIndex[index] -= 1
		log.Printf("Server %v has wrong entry at %v, continue..", index, state.State.VolatileLeaderState.NextIndex[index])
		state.State.Mu.Unlock()
		time.Sleep(time.Millisecond * (state.HeartbeatTimeoutMin / 4))
	}
}

func ReplicateLog(times int) bool {
	ch := make(chan bool)
	for i := 0; i < state.ServerNum; i += 1 {
		if i == state.Me {
			continue
		}
		go ReplicateLogOne(i, ch, times)
	}
	for i := 1; i*2 <= state.ServerNum; i++ {
		b := <-ch
		if !b {
			return false
		}
	}
	return true
}

func initLeaderState(term int) error {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	if state.State.PersistentState.CurrentTerm != term {
		return errors.New("not a leader") // late
	}
	state.State.Role = state.Leader
	state.LeaderNum = state.Me

	match_index := make(map[int]int)
	next_index := make(map[int]int)
	for i := 0; i < state.ServerNum; i += 1 {
		match_index[i] = 0
		next_index[i] = len(state.State.PersistentState.Log)
	}
	state.State.VolatileLeaderState = state.VolatileLeaderState{NextIndex: next_index, MatchIndex: match_index}

	return nil
}

func makeHeartBeat() {
	for {
		if !ReplicateLog(4) {
			return // not a leader now
		}
		time.Sleep(time.Millisecond * (state.HeartbeatTimeoutMin / 2))
	}
}

func BecomeLeader(term int) {
	log.Printf("Become leader at term %v\n", term)
	if initLeaderState(term) == nil {
		go makeHeartBeat()
	}
}
