package main

import (
	"fmt"
	"reflect"
)

type Handler func(string) string

type Transition struct {
	from         string
	toSuccess    string
	toFailure    string
	branch       bool
	waitForEvent bool
	eventName    string
}

type FSM struct {
	StartState  string
	State       string
	Handlers    map[string]string
	Transitions []Transition
}

func (fsm *FSM) AddState(stateName string, handlerFunc string) {
	fsm.Handlers[stateName] = handlerFunc
}

func (fsm *FSM) SendEvent(eventName string, eventParam string) error {
	// Find the transition that matches the state/event
	for _, t := range fsm.Transitions {
		if t.from == fsm.State && t.eventName == eventName {
			fmt.Println("Transition:", t)
			// Trigger action
			action, stateFound := fsm.Handlers[fsm.State]
			if !stateFound {
				return fmt.Errorf("Error: State '%s' not found", fsm.State)
			}
			nextState := fsm.callAction(action, eventParam)

			if t.branch {
				fsm.State = nextState
			} else {
				fsm.State = t.toSuccess
			}
			return nil
		}
	}
	return fmt.Errorf("Error: No transition supports the given state/event combination - %s/%s", fsm.State, eventName)
}

// callAction uses reflection to call an action using its name
func (fsm *FSM) callAction(action string, arg string) string {
	obj := reflect.ValueOf(fsm)
	method := obj.MethodByName(action)
	value := method.Call([]reflect.Value{reflect.ValueOf(arg)})[0]
	return value.Interface().(string)
}

/**** Local package extentions ***/

func (fsm *FSM) Log(param string) string {
	fmt.Println(param)
	return "fsm.transition.to.success"
}

func main() {
	fsm := FSM{
		StartState:  "DISARMED",
		State:       "DISARMED",
		Handlers:    map[string]string{},
		Transitions: []Transition{},
	}

	fsm.AddState("DISARMED", "Log")
	fsm.AddState("ENTER_CODE", "Log")

	t := Transition{
		from:         "DISARMED",
		toSuccess:    "ENTER_CODE",
		branch:       false,
		waitForEvent: true,
		eventName:    "ARM",
	}
	fsm.Transitions = append(fsm.Transitions, t)

	err := fsm.SendEvent("ARM", "param")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Current state: ", fsm.State)
}
