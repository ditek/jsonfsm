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
func (fsm *FSM) SetState(name string) error {
	newState, err := fsm.GetState(name)
	if err != nil {
		return err
	}
	fsm.CurrentState = newState
	fmt.Println("SetState:", newState)
	if fsm.CurrentState.WaitForEvent {
		return nil
	}

	// The state doesn't wait for an event so perform next transition
	// Find the transition that matches the state
	for _, t := range fsm.Transitions {
		if t.From == fsm.CurrentState.Name {
			fsm.beginTransition(t, fsm.CurrentState.ActionArg)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the current state - '%s'", fsm.CurrentState.Name)
}

// SendEvent sends a new event to the state machine
// Takes event name and a parameter to be passed to the action
// Returns an error if the state/event combination is not found
func (fsm *FSM) SendEvent(eventName string, eventParam string) error {
	// Find the transition that matches the state/event
	fmt.Println("SendEvent:", eventName, eventParam)
	for _, t := range fsm.Transitions {
		if t.From == fsm.CurrentState.Name && t.Event == eventName {
			fsm.beginTransition(t, eventParam)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the current state ('%s') and the sent event ('%s')", fsm.CurrentState.Name, eventName)
}

// beginTransition begins a new transition
// Returns an error if the state is not found
func (fsm *FSM) beginTransition(t Transition, actionArg string) error {
	fmt.Println("beginTransition: actionArg =", actionArg, t)
	success := fsm.callAction(actionArg)

	// Choose the next state depending on the action returned
	// value and whether the transition supports branching
	var nextState string
	if t.Branch && !success {
		nextState = t.ToFailure
	} else {
		nextState = t.ToSuccess
	}

	return fsm.SetState(nextState)
}

// callAction uses reflection to call an action using its name
func (fsm *FSM) callAction(arg string) bool {
	obj := reflect.ValueOf(fsm)
	method := obj.MethodByName(fsm.CurrentState.Action)
	value := method.Call([]reflect.Value{reflect.ValueOf(arg)})[0]
	return value.Interface().(bool)
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
func (fsm *FSM) Log(arg string) bool {
	fmt.Println(arg)
	return true
}

// ValidateCode checks the received code against the expected one
func (fsm *FSM) ValidateCode(code string) bool {
	return code == fsm.ExpectedCode
}

// SendResponse send and http response
func (fsm *FSM) SendResponse(response string) bool {
	fmt.Println("SendResponse: ", response)
	if response == "OK" {
		fmt.Println("Response is OK")
	} else {
		fmt.Println("Response is ERROR")
	}
	return true
}

/**** REST End Points ****/

// Event represents a received HTTP event
type Event struct {
	Action string `json:"action"`
	Param  string `json:"param"`
}

func eventHandler(w http.ResponseWriter, r *http.Request, fsm *FSM) {
	defer r.Body.Close()
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := fsm.SendEvent(event.Action, event.Param)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println("Current state: ", fsm.CurrentState.Name)
	respondWithJSON(w, http.StatusCreated, event)
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
	file, err := os.Open("fsm.json")
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

	fsm.SetState(fsm.InitialState)

	r := mux.NewRouter()
	r.HandleFunc("/send_event", func(w http.ResponseWriter, r *http.Request) {
		eventHandler(w, r, &fsm)
	}).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}

	// err = fsm.SendEvent("ARM", "Arming the machine!")
	// err = fsm.SendEvent("USER_CODE", "123")
	// fmt.Println("Current state: ", fsm.CurrentState.Name)
}
