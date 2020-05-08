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

	subscribe()
}

func subscribe() {

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
