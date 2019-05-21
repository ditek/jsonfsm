package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ditek/jsonfsm/gofsm"
	"github.com/gorilla/mux"
)

// FSM is a local alias to allow type extension
type FSM gofsm.FSM

const expectedCode = "123"

var httpWriter http.ResponseWriter

// Event represents a received HTTP event
type Event struct {
	Action string `json:"action"`
	Param  string `json:"param"`
}

/**** REST End Points and Functions ****/

func eventHandler(w http.ResponseWriter, r *http.Request, fsm *gofsm.FSM) {
	defer r.Body.Close()
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	httpWriter = w
	err := fsm.SendEvent(event.Action, event.Param)
	if err != nil {
		log.Println(err)
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
}

/****** Event Handlers *******/

// logRespond logs the received arg and sends an HTTP success response
func logRespond(arg string) bool {
	RespondWithJSON(httpWriter, http.StatusOK, arg)
	log.Println(arg)
	return true
}

// logArg logs the received arg
func logArg(arg string) bool {
	log.Println(arg)
	return true
}

// validateCode checks the received code against the expected one
func validateCode(code string) bool {
	return code == expectedCode
}

// sendResponse send an http response based on the passed argument
func sendResponse(arg string) bool {
	if arg == "OK" {
		RespondWithJSON(httpWriter, http.StatusOK, "CODE OK")
	} else {
		RespondWithError(httpWriter, http.StatusNotAcceptable, "WRONG CODE")
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
	fsm := gofsm.FSM{}
	err = json.Unmarshal(data, &fsm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fsm)

	// Initialize the state machine and register event handlers
	fsm.Init()
	fsm.Register("Log", logArg)
	fsm.Register("LogRespond", logRespond)
	fsm.Register("ValidateCode", validateCode)
	fsm.Register("SendResponse", sendResponse)

	r := mux.NewRouter()
	r.HandleFunc("/send_event", func(w http.ResponseWriter, r *http.Request) {
		eventHandler(w, r, &fsm)
	}).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
