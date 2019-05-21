package gofsm

import (
	"fmt"
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

// Handler defines the signature the event handler function must have
type Handler func(string) bool

// FSM represents the state machine
type FSM struct {
	InitialState string       `json:"initialState"`
	States       []State      `json:"states"`
	CurrentState State        `json:"-"`
	Transitions  []Transition `json:"transitions"`
	ExpectedCode string       `json:"expectedCode"`
	handlers     map[string]Handler
}

// Init initializes the state machine
func (fsm *FSM) Init() {
	fsm.SetState(fsm.InitialState)
	fsm.handlers = map[string]Handler{}
}

// Register registers an event handler
func (fsm *FSM) Register(name string, f Handler) {
	fsm.handlers[name] = f
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
	fmt.Println("Current state: ", fsm.CurrentState.Name)
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
	// fmt.Println("SendEvent:", eventName, eventParam)
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
	// fmt.Println("beginTransition: actionArg =", event.Param, t)
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
func (fsm *FSM) callAction(actionArg string) bool {
	handler := fsm.handlers[fsm.CurrentState.Action]
	return handler(actionArg)
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
