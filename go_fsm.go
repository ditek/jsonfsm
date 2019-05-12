package main

import (
	"fmt"
	"reflect"
)

// Transition represents an FSM transition
type Transition struct {
	from      string
	toSuccess string
	toFailure string
	branch    bool
	eventName string
}

// State epresents an FSM state
type State struct {
	name         string
	action       string
	actionArg    string
	waitForEvent bool
}

// FSM represents the state machine
type FSM struct {
	StartState   string
	States       map[string]State
	CurrentState State
	Transitions  []Transition
	ExpectedCode string
}

// AddState adds a new state to the state machine
func (fsm *FSM) AddState(stateName string, action string,
	actionArg string, waitForEvent bool) {
	fsm.States[stateName] = State{
		name:         stateName,
		action:       action,
		actionArg:    actionArg,
		waitForEvent: waitForEvent,
	}
}

// SendEvent sends a new event to the state machine
// Takes event name and a parameter to be passed to the action
// Returns an error if the state/event combination is not found
func (fsm *FSM) SendEvent(eventName string, eventParam string) error {
	// Find the transition that matches the state/event
	fmt.Println("SendEvent:", eventName, eventParam)
	for _, t := range fsm.Transitions {
		if t.from == fsm.CurrentState.name && t.eventName == eventName {
			fsm.beginTransition(t, eventParam)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the given state/event combination - %s/%s", fsm.CurrentState.name, eventName)
}

// SetState sets the state machine to the specified state
// Returns an error if the state is not found
func (fsm *FSM) SetState(name string) error {
	newState, found := fsm.States[name]
	if !found {
		return fmt.Errorf("Error: State '%s' not found in states list", name)
	}
	fsm.CurrentState = newState
	fmt.Println("SetState:", newState)
	if fsm.CurrentState.waitForEvent {
		return nil
	}

	// The state doesn't wait for an event so perform next transition
	// Find the transition that matches the state
	for _, t := range fsm.Transitions {
		if t.from == fsm.CurrentState.name {
			fsm.beginTransition(t, fsm.CurrentState.actionArg)
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the given state combination - %s", fsm.CurrentState.name)
}

// beginTransition begins a new transition
// Returns an error if the state is not found
func (fsm *FSM) beginTransition(t Transition, actionArg string) error {
	fmt.Println("beginTransition: actionArg =", actionArg, t)
	success := fsm.callAction(actionArg)

	// Choose the next state depending on the action returned
	// value and whether the transition supports branching
	var nextState string
	if t.branch && !success {
		nextState = t.toFailure
	} else {
		nextState = t.toSuccess
	}

	return fsm.SetState(nextState)
}

// callAction uses reflection to call an action using its name
func (fsm *FSM) callAction(arg string) bool {
	obj := reflect.ValueOf(fsm)
	method := obj.MethodByName(fsm.CurrentState.action)
	value := method.Call([]reflect.Value{reflect.ValueOf(arg)})[0]
	return value.Interface().(bool)
}

// New creates and initializes a new state machine
func New(startState string, expectedCode string) *FSM {
	fsm := &FSM{
		StartState:   startState,
		States:       map[string]State{},
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
	fsm := New("DISARMED", "123")
	fsm.AddState("DISARMED", "Log", "", true)
	fsm.AddState("ENTER_CODE", "ValidateCode", "", true)
	fsm.AddState("SEND_OK_RESPONSE", "SendResponse", "OK", false)
	fsm.AddState("SEND_ERROR_RESPONSE", "SendResponse", "ERROR", false)
	fsm.AddState("ARMED", "Log", "", true)
	fsm.SetState(fsm.StartState)

	t := Transition{
		from:      "DISARMED",
		toSuccess: "ENTER_CODE",
		branch:    false,
		eventName: "ARM",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	t = Transition{
		from:      "ENTER_CODE",
		toSuccess: "SEND_OK_RESPONSE",
		toFailure: "SEND_ERROR_RESPONSE",
		branch:    true,
		eventName: "USER_CODE",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	t = Transition{
		from:      "SEND_OK_RESPONSE",
		toSuccess: "ARMED",
		branch:    false,
		eventName: "FOREVER",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	fmt.Println()

	err := fsm.SendEvent("ARM", "Arming the machine!")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current state: ", fsm.CurrentState.name)
	fmt.Println()

	err = fsm.SendEvent("USER_CODE", "123")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current state: ", fsm.CurrentState.name)
}
