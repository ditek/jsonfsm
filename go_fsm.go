package main

import (
	"fmt"
)

type Handler func(string) string

type Event struct {
	name string
	// State to be affected by event
	state string
}

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
	Handlers    map[string]Handler
	Transitions map[Event]Transition
}

func (fsm *FSM) AddState(stateName string, handlerFunc Handler) {
	fsm.Handlers[stateName] = handlerFunc
}

func (fsm *FSM) SendEvent(eventName string, eventParam string) error {
	eventObj := Event{
		name:  eventName,
		state: fsm.State,
	}
	// Check if the current state accepts the event
	if transition, found := fsm.Transitions[eventObj]; found {
		fmt.Println("Transition:", transition)
		// Trigger action
		action, stateFound := fsm.Handlers[fsm.State]
		if !stateFound {
			return fmt.Errorf("Error: State '%s' not found", fsm.State)
		}
		nextState := action(eventParam)
		if transition.branch {
			fsm.State = nextState
		} else {
			fsm.State = transition.toSuccess
		}
		return nil
	}
	return fmt.Errorf("Error: event/state combination not supported")
}

/**** Local package extentions ***/

func (fsm *FSM) Log(param string) string {
	fmt.Println(param)
	return "fsm.transition.to.sucess"
}

func main() {
	x := FSM{
		StartState:  "DISARMED",
		State:       "DISARMED",
		Handlers:    map[string]Handler{},
		Transitions: map[Event]Transition{},
	}

	x.AddState("DISARMED", x.Log)
	x.AddState("ENTER_CODE", x.Log)

	e := Event{
		name:  "ARM",
		state: "DISARME",
	}
	t := Transition{
		from:         "DISARMED",
		toSuccess:    "ENTER_CODE",
		branch:       false,
		waitForEvent: true,
		eventName:    "ARM",
	}
	x.Transitions[e] = t

	err := x.SendEvent("ARM", "param")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Current state: ", x.State)
}
