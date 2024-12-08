package raft_state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type VolatileState struct {
	CommitIndex int
	LastApplied int
}

type VolatileLeaderState struct {
	NextIndex  map[int]int
	MatchIndex map[int]int
}

const (
	Create = iota
	Read   = iota
	Update = iota
	Delete = iota
)

type Log struct {
	Uuid string `json:"uuid"`
	Op   string `json:"op"`
	Body string `json:"body"`
	Term int    `json:"term"`
}

type PersistentState struct {
	CurrentTerm int   `json:"currentTerm"`
	VotedFor    int   `json:"votedFor"`
	Log         []Log `json:"log"`
}

const (
	Follower  = iota
	Leader    = iota
	Candidate = iota
)

type MainState struct {
	VolatileState       VolatileState
	VolatileLeaderState VolatileLeaderState
	PersistentState     PersistentState
	Role                int
	Mu                  sync.Mutex
}

// lock is already acquired
func WritePersState(state PersistentState) {
	json, err := json.Marshal(state)
	if err != nil {
		log.Fatalf("WritePersState failed: %s", err)
	}
	err = os.WriteFile("pers_state.json", json, 0666)
	if err != nil {
		log.Fatalf("WritePersState failed: %s", err)
	}
	// log.Printf("Written to file: %+v", state)
}

var State MainState
var HeartbeatChan chan struct{}

var Me int = 0
var ServerNum int = 0
var Num2server_raft map[int]string
var Num2server_http map[int]string
var LeaderNum int

const HeartbeatTimeoutMin = 2000
const HeartbeatTimeoutMax = 4000

func InitEnv() {
	me, err := strconv.Atoi(os.Getenv("ME"))
	if err != nil {
		fmt.Println("Error: ME not set or invalid")
		os.Exit(1)
	}
	Me = me

	serverNum, err := strconv.Atoi(os.Getenv("SERVER_NUM"))
	if err != nil {
		fmt.Println("Error: SERVER_NUM not set or invalid")
		os.Exit(1)
	}
	ServerNum = serverNum

	Num2server_raft = make(map[int]string)
	servers_raft := os.Getenv("SERVERS_RAFT")
	for i, addr := range strings.Split(servers_raft, ",") {
		Num2server_raft[i] = addr
	}

	Num2server_http = make(map[int]string)
	servers_http := os.Getenv("SERVERS_HTTP")
	for i, addr := range strings.Split(servers_http, ",") {
		Num2server_http[i] = addr
	}
}
func InitState() {
	InitEnv()
	persstate := PersistentState{}
	persstate.Log = make([]Log, 0)
	persstate.VotedFor = -1
	bytes, err := os.ReadFile("pers_state.json")
	if err == nil {
		json.Unmarshal(bytes, &persstate)
		log.Printf("Readed from json: %+v", persstate)
	}

	State.PersistentState = persstate
	State.Role = Follower
}

func CheckIfLeader() (bool, string) {
	State.Mu.Lock()
	defer State.Mu.Unlock()
	if Me == LeaderNum {
		return true, ""
	}
	return false, Num2server_http[LeaderNum]
}
