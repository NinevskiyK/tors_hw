package raft_main

import (
	"log"
	"raft/raft_request"
	state "raft/raft_state"
	"sync"
	"time"
)

func StartElection() {
	current_term := 0
	{
		state.State.Mu.Lock()

		if state.State.Role == state.Leader {
			state.State.Mu.Unlock()
			return
		}

		state.State.Role = state.Candidate
		current_term = state.State.PersistentState.CurrentTerm + 1

		state.State.PersistentState.CurrentTerm += 1
		state.State.PersistentState.VotedFor = state.Me
		state.WritePersState(state.State.PersistentState)

		state.State.Mu.Unlock()
	}
	log.Printf("Started election with term %v\n", current_term)

	var voted = make(map[int]bool)
	voted[state.Me] = true
	var max_term int
	var mu sync.Mutex

	for {
		var wg sync.WaitGroup
		max_term = 0
		for i := 0; i < state.ServerNum; i++ {
			if i != state.Me {
				wg.Add(1)
				go func(j int) {
					defer wg.Done()
					resp, err := raft_request.RequestVote(state.Num2server_raft[i])

					if err != nil {
						return // not a candidate
					}

					mu.Lock()
					defer mu.Unlock()
					if resp.VoteGranted {
						voted[j] = true
					}
					if max_term < resp.Term {
						max_term = resp.Term
					}
				}(i)
			}
		}
		wg.Wait()
		{
			state.State.Mu.Lock()
			if current_term < state.State.PersistentState.CurrentTerm || current_term < max_term {
				if max_term > state.State.PersistentState.CurrentTerm {
					ConvertToFollower(max_term)
				}
				state.State.Mu.Unlock()
				return
			}
			state.State.Mu.Unlock()
		}
		voted_count := 0
		for _, v := range voted {
			if v {
				voted_count += 1
			}
		}
		log.Printf("voted num: %v in term: %v", voted_count, current_term)
		if voted_count*2 > state.ServerNum {
			BecomeLeader(current_term)
			return
		}
		time.Sleep(time.Millisecond * (state.HeartbeatTimeoutMin / 4))
	}
}
