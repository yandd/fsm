FSM - Finite state machine in Go
===========================

## usage
```sh
go get github.com/yandd/fsm
```

## example
![fsm](fsm.png)

```go
f, err := fsm.NewFSM("inited", []interface{}{"ended"}, fsm.FSMEvents{
	{Name: "start", From: "inited", To: "started"},
	{Name: "work", From: "started", To: "working"},
	{Name: "end", From: []interface{}{"started", "working"}, To: "ended"},
})
```
