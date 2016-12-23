package fsm

import (
	"testing"

	"encoding/json"

	"log"
)

func toJson(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "Err: " + err.Error()
	}

	return string(data)
}

func TestNextState(t *testing.T) {
	f, err := NewFSM(nil, []interface{}{}, FSMEvents{
		{Name: "start", From: "inited", To: "started"},
		{Name: "work", From: "started", To: "working"},
		{Name: "end", From: []interface{}{"started", "working"}, To: "ended"},
	})
	if err != nil {
		t.Error(err)
		return
	}

	next, err := NextState(f, "started", "work")
	log.Println(f.Graph())
	log.Println(next, err)

	log.Println(f.Dot("test"))
}

