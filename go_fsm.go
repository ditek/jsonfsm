package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	return fmt.Errorf("Error: No transition supports the given state combination - %s", fsm.CurrentState.Name)
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
	return fmt.Errorf("Error: No transition supports the given state/event combination - %s/%s", fsm.CurrentState.Name, eventName)
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

func (fsm *FSM) Log(arg string) bool {
	fmt.Println(arg)
	return true
}

func (fsm *FSM) ValidateCode(code string) bool {
	return code == fsm.ExpectedCode
}

func (fsm *FSM) SendResponse(response string) bool {
	fmt.Println("SendResponse: ", response)
	if response == "OK" {
		fmt.Println("Response is OK")
	} else {
		fmt.Println("Response is ERROR")
	}
	return true
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

	var fsm FSM
	err = json.Unmarshal(data, &fsm)
	if err != nil {
		log.Fatal(err)
	}

	fsm.SetState(fsm.InitialState)

	err = fsm.SendEvent("ARM", "Arming the machine!")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current state: ", fsm.CurrentState.Name)
	fmt.Println()

	err = fsm.SendEvent("USER_CODE", "123")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current state: ", fsm.CurrentState.Name)
}
