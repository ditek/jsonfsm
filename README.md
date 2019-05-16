# JSON FSM
Finite State Machine for Go with JSON representation.

## Usage

### Setup
You need Go version 1.11 or newer to run the project according to these instructions.

After cloning the project run:

```sh
go build
```

Run the executable with the JSON file that describes the state machine as an argument:

```sh
./jsonfsm fsm.json
```

If things go well, you should see a log message specifying the current state.

### Sending Events
Events are sent as HTTP POST requests and have a body that follows this format.

```json
{
    "action": "action_name",
    "param": "action_param"
}
```
The given example expects requests on `localhost:3000/send_event`.

An error message will be printed if the current state does not support the given event. This is a sample output of the script.

```sh
2019/05/15 10:26:01 Current state:  DISARMED
2019/05/15 10:26:05 This was sent with the ARM event
2019/05/15 10:26:05 Current state:  ENTER_CODE
2019/05/15 11:04:05 Error: No transition supports the current state ('ENTER_CODE') and the sent event ('ARM')
```

### JSON File Format
The JSON file should follow the following format.

```
{
    "initialState": "STATE1",     // Initial FSM state
    "expectedCode": "123",          // Code to check against to determine transition destination
    "states": [
        {
            "name": "STATE1",
            "action": "Log",        // The action that is triggered by the state
            "waitForEvent": true,   // Whether the state should wait for an event or transition immediately
            "sendResponse": true    // Whether the state action should send a response 
        },
        {
            "name": "STATE2",
            ...
        }
    ],
    // List of supported transitions
    "transitions": [
        {
            "from": "STATE1",
            "branch": true,         // Whether we branch based on the boolean provide by the source state
            "toSuccess": "STATE2",  // Next state on success
            "toFailure": "STATE3",  // Next state on failure
            "event": "USER_CODE"    // The event that triggers the transition
        },
        {
            "from": "STATE2",
            "branch": false,
            "toSuccess": "STATE4",  // 'toFailure' state not needed if we don't branch
            "event": "ARM"
        }
    ],
    // List of supported events
    "events": [
        "ARM",
        "USER_CODE"
    ]
}
```

## Notes
`fsm.Init()` needs to be called after creating the FSM instance.
