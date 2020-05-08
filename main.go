package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"go.i3wm.org/i3/v4"
)

func main() {

	//i3 overrides to work with sway
	i3.SocketPathHook = func() (string, error) {
		out, err := exec.Command("sway", "--get-socketpath").CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("getting sway socketpath: %v (output: %s)", err, out)
		}
		return string(out), nil
	}

	i3.IsRunningHook = func() bool {
		out, err := exec.Command("pgrep", "-c", "-f", "/usr/bin/sway").CombinedOutput()
		if err != nil {
			log.Printf("sway running: %v (output: %s)", err, out)
		}
		return bytes.Compare(bytes.TrimSpace(out), []byte("1")) == 0
	}

	sp, _ := i3.SocketPathHook()
	ir := i3.IsRunningHook()
	fmt.Println("foo", sp, ir)

	subscription := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType, i3.ShutdownEventType)

	o := NewOrg()

	wsse, err := i3.GetWorkspaces()
	for _, w := range wsse {
		o.WorkspaceFront(w.ID)
	}
	

	for n := subscription.Next(); n; n = subscription.Next() {
		event := subscription.Event()
		switch event.(type) {
		case *i3.ShutdownEvent:
			fmt.Println("shutting down")
			break
		case *i3.WindowEvent:
			ev := event.(*i3.WindowEvent)
			change := ev.Change

			fmt.Println(change, ev.Container.ID)
		case *i3.WorkspaceEvent:
			ev := event.(*i3.WorkspaceEvent)

			fmt.Println(ev.Change, ev.Current, ev.Old)
			switch ev.Change {
			case "init":
				o.WorkspaceFront(i3.WorkspaceID(ev.Current.ID))
			}
		}
	}

	err = subscription.Close()
	if err != nil {
		fmt.Println(err)
	}

}

type org struct {
	m map[i3.WorkspaceID]map[i3.NodeID]i3.Node
	w []i3.WorkspaceID
	n []i3.NodeID
}

func NewOrg() *org {
	return &org{
		m: make(map[i3.WorkspaceID]map[i3.NodeID]i3.Node, 0),
	}
}

func (o *org) WorkspaceFront(wsid i3.WorkspaceID) {

	// If new, create new window cache
	if _, ok := o.m[wsid]; !ok {
		o.m[wsid] = make(map[i3.NodeID]i3.Node, 0)
	}

	// Move Workspace to front
	for i, id := range o.w {
		if id == wsid {
			o.w = append(o.w[:i], o.w[i+1:]...)
		}
	}
	o.w = append([]i3.WorkspaceID{wsid}, o.w...)
}
