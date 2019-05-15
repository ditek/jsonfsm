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

/**** REST End Points and Functions ****/

func eventHandler(w http.ResponseWriter, r *http.Request, fsm *gofsm.FSM) {
	defer r.Body.Close()
	var event gofsm.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		gofsm.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	event.Writer = w
	err := fsm.SendEvent(event)
	if err != nil {
		log.Println(err)
		gofsm.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
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

	// Initialize the state machine
	fsm.Init()

	r := mux.NewRouter()
	r.HandleFunc("/send_event", func(w http.ResponseWriter, r *http.Request) {
		eventHandler(w, r, &fsm)
	}).Methods("POST")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
