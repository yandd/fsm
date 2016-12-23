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

## support export .dot file
```go
f.Dot("diag_name")
```

```dot
digraph diag_name {
	rankdir=LR;
	node [shape = doublecircle];ended ;
	node [shape = circle];

	inited -> started [ label = "start" ];
	started -> working [ label = "work" ];
	started -> ended [ label = "end" ];
	working -> ended [ label = "end" ];
}
```
