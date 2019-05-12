package main

import (
	"fmt"
	"reflect"
)

const expectedCode = "123"

type Transition struct {
	from      string
	toSuccess string
	toFailure string
	branch    bool
	eventName string
}

type State struct {
	name         string
	action       string
	actionArg    string
	waitForEvent bool
}

type FSM struct {
	StartState   string
	States       map[string]State
	CurrentState State
	Transitions  []Transition
}

func (fsm *FSM) AddState(stateName string, action string,
	actionArg string, waitForEvent bool) {
	fsm.States[stateName] = State{
		name:         stateName,
		action:       action,
		actionArg:    actionArg,
		waitForEvent: waitForEvent,
	}
}

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

func (fsm *FSM) setState(name string) error {
	newState, found := fsm.States[name]
	if !found {
		return fmt.Errorf("Error: State '%s' not found in states list", name)
	}
	fsm.CurrentState = newState
	fmt.Println("setState:", newState)
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

	fsm.setState(nextState)
	return nil
}

// callAction uses reflection to call an action using its name
func (fsm *FSM) callAction(arg string) bool {
	obj := reflect.ValueOf(fsm)
	method := obj.MethodByName(fsm.CurrentState.action)
	value := method.Call([]reflect.Value{reflect.ValueOf(arg)})[0]
	return value.Interface().(bool)
}

/**** Local package extentions ***/

func (fsm *FSM) Log(arg string) bool {
	fmt.Println(arg)
	return true
}

func (fsm *FSM) ValidateCode(code string) bool {
	return code == expectedCode
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
	fsm := FSM{
		StartState:  "DISARMED",
		States:      map[string]State{},
		Transitions: []Transition{},
	}

	fsm.AddState("DISARMED", "Log", "", true)
	fsm.AddState("ENTER_CODE", "ValidateCode", "", true)
	fsm.AddState("SEND_OK_RESPPONSE", "SendResponse", "OK", false)
	fsm.AddState("SEND_ERROR_RESPPONSE", "SendResponse", "ERROR", false)
	fsm.AddState("ARMED", "Log", "", true)
	fsm.setState(fsm.StartState)

	t := Transition{
		from:      "DISARMED",
		toSuccess: "ENTER_CODE",
		branch:    false,
		eventName: "ARM",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	t = Transition{
		from:      "ENTER_CODE",
		toSuccess: "SEND_OK_RESPPONSE",
		toFailure: "SEND_ERROR_RESPPONSE",
		branch:    true,
		eventName: "USER_CODE",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	t = Transition{
		from:      "SEND_OK_RESPPONSE",
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
