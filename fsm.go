package fsm

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sort"
)

var (
	ErrFSMEventsEmpty          = errors.New("<FSM> events empty")
	ErrFSMEventsConflict       = errors.New("<FSM> events conflict")
	ErrFSMParamInvalid         = errors.New("<FSM> param invalid")
	ErrFSMEventNotFound        = errors.New("<FSM> event not found")
	ErrFSMStateInEventNotFound = errors.New("<FSM> state in event not found")
)

type Comparator interface {
	Compare(interface{}) int
}

type FSMEvent struct {
	Name interface{}
	From interface{}
	To   interface{}
}

type FSMEvents []FSMEvent

func (e FSMEvents) Len() int           { return len(e) }
func (e FSMEvents) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e FSMEvents) Less(i, j int) bool {
	_, nameOk := e[i].Name.(Comparator);
	_, fromOk := e[i].From.(Comparator);
	_, toOk := e[i].To.(Comparator);

	if nameOk && e[i].Name.(Comparator).Compare(e[j].Name) < 0 {
		return true
	} else if !nameOk || (nameOk && e[i].Name.(Comparator).Compare(e[j].Name) == 0) {
		if fromOk && e[i].From.(Comparator).Compare(e[j].From) < 0 {
			return true
		} else if fromOk && e[i].From.(Comparator).Compare(e[j].From) == 0 {
			if toOk && e[i].To.(Comparator).Compare(e[j].To) < 0 {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

type FSMGraph map[interface{}]map[interface{}]interface{}

type FSM struct {
	mutex sync.RWMutex

	initial interface{}
	accepts map[interface{}]interface{}
	current interface{}
	graph   FSMGraph
}

func (f *FSM) buildGraph(events FSMEvents) error {
	if len(events) == 0 {
		return ErrFSMEventsEmpty
	}

	f.graph = make(FSMGraph)

	for _, event := range events {
		_, ok := f.graph[event.Name]
		if !ok {
			f.graph[event.Name] = make(map[interface{}]interface{})
		}

		typ := reflect.TypeOf(event.From)
		if typ.Kind() == reflect.Slice {
			for _, v := range event.From.([]interface{}) {
				to, ok := f.graph[event.Name][v]
				if ok {
					if !reflect.DeepEqual(to, event.To) {
						return errors.New(fmt.Sprintln(ErrFSMEventsConflict, event.Name, v))
					}
				} else {
					f.graph[event.Name][v] = event.To
				}
			}
		} else {
			to, ok := f.graph[event.Name][event.From]
			if ok {
				if !reflect.DeepEqual(to, event.To) {
					return errors.New(fmt.Sprintln(ErrFSMEventsConflict, event.Name, event.From))
				}
			}
			f.graph[event.Name][event.From] = event.To
		}
	}

	return nil
}

func NewFSM(initial interface{}, accepts []interface{}, events FSMEvents) (*FSM, error) {
	f := &FSM{
		initial: initial,
	}

	if err := f.buildGraph(events); err != nil {
		return nil, err
	}

	if len(accepts) > 0 {
		f.accepts = make(map[interface{}]interface{}, len(accepts))
		for _, state := range accepts {
			f.accepts[state] = nil
		}
	}

	f.current = f.initial

	return f, nil
}

func (f *FSM) Reset() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.current = f.initial
}

func (f *FSM) Acceptable() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if len(f.accepts) == 0 {
		return false
	}

	if _, ok := f.accepts[f.current]; ok {
		return true
	}

	return false
}

func (f *FSM) SetCurrent(state interface{}) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.current = state
}

func (f *FSM) GetCurrent() interface{} {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.current
}

func (f *FSM) Next(evName interface{}) (interface{}, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if evName == nil {
		return nil, ErrFSMParamInvalid
	}

	_, ok := f.graph[evName]
	if !ok {
		return nil, ErrFSMEventNotFound
	}

	nextState, ok := f.graph[evName][f.current]
	if !ok {
		return nil, ErrFSMStateInEventNotFound
	}

	f.current = nextState

	return nextState, nil
}

func (f *FSM) Graph() FSMGraph {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.graph
}

func (f *FSM) Dot(name string) string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	buf := bytes.Buffer{}

	buf.WriteString(fmt.Sprintf(`
digraph %s {
	rankdir=LR;
	node [shape = doublecircle];`, name))

	for a, _ := range f.accepts {
		buf.WriteString(fmt.Sprintf("%v ", a))
	}

	if len(f.accepts) > 0 {
		buf.WriteString(";")
	}

	buf.WriteString("\n	node [shape = circle];\n\n")

	events := FSMEvents{}

	for e, s := range f.graph {
		for curr, next := range s {
			events = append(events, FSMEvent{e, curr, next})
			//buf.WriteString(fmt.Sprintf("	%v -> %v [ label = \"%v\" ];\n", curr, next, e))
		}
	}

	if len(events) > 0 {
		_, nameOk := events[0].Name.(Comparator);
		_, fromOk := events[0].From.(Comparator);

		if nameOk || fromOk {
			sort.Sort(events)
		}
	}

	for _, e := range events {
		buf.WriteString(fmt.Sprintf("	%v -> %v [ label = \"%v\" ];\n", e.From, e.To, e.Name))
	}

	buf.WriteString("}\n")

	return buf.String()
}

func NextState(f *FSM, currState interface{}, evName interface{}) (interface{}, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if f == nil || evName == nil {
		return nil, ErrFSMParamInvalid
	}

	_, ok := f.graph[evName]
	if !ok {
		return nil, ErrFSMEventNotFound
	}

	nextState, ok := f.graph[evName][currState]
	if !ok {
		return nil, ErrFSMStateInEventNotFound
	}

	return nextState, nil
}
