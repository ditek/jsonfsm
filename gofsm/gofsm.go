package gofsm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

// Transition represents an FSM transition
type Transition struct {
	From      string `json:"from"`
	ToSuccess string `json:"toSuccess"`
	ToFailure string `json:"toFailure,omitempty"`
	Branch    bool   `json:"branch"`
	Event     string `json:"event,omitempty"`
}

// State presents an FSM state
type State struct {
	Name         string `json:"name"`
	Action       string `json:"action"`
	ActionArg    string `json:"action_arg,omitempty"`
	WaitForEvent bool   `json:"waitForEvent"`
	SendResponse bool   `json:"sendResponse"`
}

// Event represents a received HTTP event
type Event struct {
	Action string              `json:"action"`
	Param  string              `json:"param"`
	Writer http.ResponseWriter `json:"writer,omitempty"`
}

// FSM represents the state machine
type FSM struct {
	InitialState string       `json:"initialState"`
	States       []State      `json:"states"`
	CurrentState State        `json:"omitempty"`
	Transitions  []Transition `json:"transitions"`
	ExpectedCode string       `json:"expectedCode"`
}

// Init initializes the state machine
func (fsm *FSM) Init() {
	fsm.SetState(fsm.InitialState, Event{})
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
	log.Println("Current state: ", fsm.CurrentState.Name)
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
	// fmt.Println("SendEvent:", event.Action, event.Param)
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
	callable := method.Interface().(func(string, http.ResponseWriter) bool)
	return callable(event.Param, event.Writer)
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

/******* Callable Actions ********/

// Log prints out the received message
func (fsm *FSM) Log(arg string, w http.ResponseWriter) bool {
	if fsm.CurrentState.SendResponse {
		RespondWithJSON(w, http.StatusOK, "")
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
		RespondWithJSON(w, http.StatusOK, "CODE OK")
	} else {
		RespondWithError(w, http.StatusNotAcceptable, "WRONG CODE")
	}
	return true
}

/****** Convenience Functions *******/

// RespondWithError sends an HTTP error response
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

// RespondWithJSON sends an custom HTTP response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
