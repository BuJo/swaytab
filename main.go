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

	// Initial workspace stack
	wsse, err := i3.GetWorkspaces()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, w := range wsse {
		o.WorkspaceFront(w.ID)
	}
	for _, w := range wsse {
		if w.Focused {
			o.WorkspaceFront(w.ID)
			break
		}
	}

	// Initial window stack
	tree, err := i3.GetTree()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, output := range tree.Root.Nodes {
		for _, workspace := range output.Nodes {
			for _, n := range workspace.FloatingNodes {
				o.WindowAddTo(*n, i3.WorkspaceID(workspace.ID))
			}
			for _, n := range workspace.FloatingNodes {
				o.WindowAddTo(*n, i3.WorkspaceID(workspace.ID))
			}
		}
	}

	for n := subscription.Next(); n; n = subscription.Next() {
		fmt.Println(o)
		event := subscription.Event()
		switch event.(type) {
		case *i3.ShutdownEvent:
			fmt.Println("shutting down")
			break
		case *i3.WindowEvent:
			ev := event.(*i3.WindowEvent)

			fmt.Println("win", ev.Change, ev.Container.ID)
			switch ev.Change {
			case "new":
				o.WindowAdd(ev.Container)
			case "focus":
				o.WindowFront(ev.Container)
			case "close":
				o.WindowDelete(ev.Container)
			}
		case *i3.WorkspaceEvent:
			ev := event.(*i3.WorkspaceEvent)

			fmt.Println("ws", ev.Change, ev.Current, ev.Old)
			switch ev.Change {
			case "focus":
				o.WorkspaceFront(i3.WorkspaceID(ev.Current.ID))
			case "init":
				o.WorkspaceFront(i3.WorkspaceID(ev.Current.ID))
			case "empty":
				o.WorkspaceDelete(i3.WorkspaceID(ev.Current.ID))
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
	n map[i3.WorkspaceID][]i3.NodeID
}

func NewOrg() *org {
	return &org{
		m: make(map[i3.WorkspaceID]map[i3.NodeID]i3.Node, 0),
		n: make(map[i3.WorkspaceID][]i3.NodeID, 0),
	}
}

func (o *org) WorkspaceFront(wsid i3.WorkspaceID) {

	// If new, create new window cache
	if _, ok := o.m[wsid]; !ok {
		o.m[wsid] = make(map[i3.NodeID]i3.Node, 0)
		o.n[wsid] = make([]i3.NodeID, 0)
	}

	// Move Workspace to front of stack
	newstack := []i3.WorkspaceID{wsid}
	for _, id := range o.w {
		if id != wsid {
			newstack = append(newstack, id)
		}
	}
	o.w = newstack
}

func (o *org) WorkspaceDelete(wsid i3.WorkspaceID) {

	// Delete cache
	delete(o.m, wsid)
	delete(o.n, wsid)

	// Remove from stack
	for i, id := range o.w {
		if id == wsid {
			o.w = append(o.w[:i], o.w[i+1:]...)
		}
	}
}

func (o *org) WindowFront(n i3.Node) {
	// Update cache
	wsid := o.w[0]
	o.m[wsid][n.ID] = n

	// Remove from stack
	newstack := []i3.NodeID{n.ID}
	for _, id := range o.n[wsid] {
		if id != n.ID {
			newstack = append(newstack, id)
		}
	}
	o.n[wsid] = newstack
}

func (o *org) WindowDelete(n i3.Node) {

	// Delete cache
	wsid := o.w[0]
	delete(o.m[wsid], n.ID)

	// Remove from stack
	for i, id := range o.n[wsid] {
		if id == n.ID {
			o.n[wsid] = append(o.n[wsid][:i], o.n[wsid][i+1:]...)
		}
	}
}

func (o *org) String() string {
	return fmt.Sprintf("|%v|%v", o.w, o.n[o.w[0]])
}

func (o *org) WindowAdd(n i3.Node) {
	// Update cache
	wsid := o.w[0]
	o.m[wsid][n.ID] = n

	o.n[wsid] = append(o.n[wsid], n.ID)
}

func (o *org) WindowAddTo(n i3.Node, wsid i3.WorkspaceID) {
	// Update cache
	o.m[wsid][n.ID] = n

	o.n[wsid] = append(o.n[wsid], n.ID)
}
