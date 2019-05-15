package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gorilla/mux"
)

// Transition represents an FSM transition
type Transition struct {
	From      string `json:"from"`
	ToSuccess string `json:"toSuccess"`
	ToFailure string `json:"toFailure,omitempty"`
	Branch    bool   `json:"branch"`
	Event     string `json:"event,omitempty"`
}

// State epresents an FSM state
type State struct {
	Name         string `json:"name"`
	Action       string `json:"action"`
	ActionArg    string `json:"action_arg,omitempty"`
	WaitForEvent bool   `json:"waitForEvent"`
	SendResponse bool   `json:"sendResponse"`
}

// FSM represents the state machine
type FSM struct {
	InitialState string       `json:"initialState"`
	States       []State      `json:"states"`
	CurrentState State        `json:"omitempty"`
	Transitions  []Transition `json:"transitions"`
	ExpectedCode string       `json:"expectedCode"`
}

// AddState adds a new state to the state machine
func (fsm *FSM) AddState(stateName string, action string,
	actionArg string, waitForEvent bool) {
	s := State{
		Name:         stateName,
		Action:       action,
		ActionArg:    actionArg,
		WaitForEvent: waitForEvent,
	}
	fsm.States = append(fsm.States, s)
}

// GetState returns the state with the matching name
// and an error if not found
func (fsm *FSM) GetState(name string) (State, error) {
	for _, s := range fsm.States {
		if s.Name == name {
			return s, nil
		}
	}
	return State{}, fmt.Errorf("Error: State '%s' not found in states list", name)
}

// SetState sets the state machine to the specified state
// Returns an error if the state is not found
func (fsm *FSM) SetState(name string, event Event) error {
	newState, err := fsm.GetState(name)
	if err != nil {
		return err
	}
	fsm.CurrentState = newState
	// fmt.Println("SetState:", newState)
	if fsm.CurrentState.WaitForEvent {
		return nil
	}

	// The state doesn't wait for an event so perform next transition
	// Find the transition that matches the state
	event.Param = fsm.CurrentState.ActionArg
	for _, t := range fsm.Transitions {
		if t.From == fsm.CurrentState.Name {
			fsm.beginTransition(t, event)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the current state - '%s'", fsm.CurrentState.Name)
}

// SendEvent sends a new event to the state machine
// Takes event name and a parameter to be passed to the action
// Returns an error if the state/event combination is not found
func (fsm *FSM) SendEvent(event Event) error {
	// Find the transition that matches the state/event
	fmt.Println("SendEvent:", event.Action, event.Param)
	for _, t := range fsm.Transitions {
		if t.From == fsm.CurrentState.Name && t.Event == event.Action {
			fsm.beginTransition(t, event)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the current state ('%s') and the sent event ('%s')", fsm.CurrentState.Name, event.Action)
}

// beginTransition begins a new transition
// Returns an error if the state is not found
func (fsm *FSM) beginTransition(t Transition, event Event) error {
	// fmt.Println("beginTransition: actionArg =", event.Param, t)
	success := fsm.callAction(event)

	// Choose the next state depending on the action returned
	// value and whether the transition supports branching
	var nextState string
	if t.Branch && !success {
		nextState = t.ToFailure
	} else {
		nextState = t.ToSuccess
	}

	return fsm.SetState(nextState, event)
}

// callAction uses reflection to call an action using its name
func (fsm *FSM) callAction(event Event) bool {
	obj := reflect.ValueOf(fsm)
	method := obj.MethodByName(fsm.CurrentState.Action)
	// Convert to a function with the right signature
	mCallable := method.Interface().(func(string, http.ResponseWriter) bool)
	return mCallable(event.Param, event.Writer)
}

// New creates and initializes a new state machine
func New(startState string, expectedCode string) *FSM {
	fsm := &FSM{
		InitialState: startState,
		States:       []State{},
		Transitions:  []Transition{},
		ExpectedCode: expectedCode,
	}
	return fsm
}

/**** Local package extentions ***/

// Log prints out the recieved message
func (fsm *FSM) Log(arg string, w http.ResponseWriter) bool {
	if fsm.CurrentState.SendResponse {
		respondWithJSON(w, http.StatusOK, "")
	}
	log.Println(arg)
	return true
}

// ValidateCode checks the received code against the expected one
func (fsm *FSM) ValidateCode(code string, w http.ResponseWriter) bool {
	return code == fsm.ExpectedCode
}

// SendResponse send and http response
func (fsm *FSM) SendResponse(response string, w http.ResponseWriter) bool {
	if response == "OK" {
		respondWithJSON(w, http.StatusOK, "CODE OK")
	} else {
		respondWithError(w, http.StatusNotAcceptable, "WRONG CODE")
	}
	return true
}

/**** REST End Points ****/

// Event represents a received HTTP event
type Event struct {
	Action string              `json:"action"`
	Param  string              `json:"param"`
	Writer http.ResponseWriter `json:"writer,omitempty"`
}

func eventHandler(w http.ResponseWriter, r *http.Request, fsm *FSM) {
	defer r.Body.Close()
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	event.Writer = w
	err := fsm.SendEvent(event)
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println("Current state: ", fsm.CurrentState.Name)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(fmt.Errorf("Usage: ./jsonfsm <file_name>"))
		os.Exit(1)
	}
	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Create the FSM from the json file
	var fsm FSM
	err = json.Unmarshal(data, &fsm)
	if err != nil {
		log.Fatal(err)
	}

	fsm.SetState(fsm.InitialState, Event{})

	r := mux.NewRouter()
	r.HandleFunc("/send_event", func(w http.ResponseWriter, r *http.Request) {
		eventHandler(w, r, &fsm)
	}).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
