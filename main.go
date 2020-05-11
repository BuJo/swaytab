package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
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

	if len(os.Args) > 1 {
		go subscribe()

		app, _ := gtk.ApplicationNew("com.github.BuJo.swaytab", glib.APPLICATION_FLAGS_NONE)
		app.Connect("activate", gtkActivated)
		app.Run(nil)
	} else {
		subscribe()
	}
}

type mode int

const (
	CYCLE mode = iota
	TOGGLE
)

func subscribe() {

	subscription := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType, i3.ShutdownEventType, i3.TickEventType)

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
			for _, n := range workspace.Nodes {
				o.WindowAddTo(*n, i3.WorkspaceID(workspace.ID))
			}
		}
	}
	for _, w := range o.m {
		for _, n := range w {
			if n.Focused {
				o.WindowFront(n)
			}
		}
	}

	press := time.Now()
	currentMode := TOGGLE
	currentWindow := 0
	var currentDeadline *time.Timer

	for n := subscription.Next(); n; n = subscription.Next() {
		event := subscription.Event()
		switch event.(type) {
		case *i3.ShutdownEvent:
			break
		case *i3.WindowEvent:
			ev := event.(*i3.WindowEvent)

			switch ev.Change {
			case "new":
				o.WindowAdd(ev.Container)
			case "focus":
				fmt.Println("win", ev.Change, o)
				switch currentMode {
				case TOGGLE:
					o.WindowFront(ev.Container)
				case CYCLE:
					// do not modify stack
					if currentDeadline == nil {
						fmt.Println("dead", "init", o)
						currentDeadline = time.AfterFunc(1000*time.Millisecond, func() {
							fmt.Println("dead", "fire", o)
							o.WindowFront(ev.Container)
							currentDeadline = nil
							currentMode = TOGGLE
						})
					} else {
						fmt.Println("dead", "reset", o)
						currentDeadline.Reset(1000 * time.Millisecond)
					}
				}
			case "close":
				o.WindowDelete(ev.Container)
			}
		case *i3.WorkspaceEvent:
			ev := event.(*i3.WorkspaceEvent)

			switch ev.Change {
			case "focus":
				o.WorkspaceFront(i3.WorkspaceID(ev.Current.ID))
			case "init":
				o.WorkspaceFront(i3.WorkspaceID(ev.Current.ID))
			case "empty":
				o.WorkspaceDelete(i3.WorkspaceID(ev.Current.ID))
			}
		case *i3.TickEvent:
			ev := event.(*i3.TickEvent)
			if strings.HasPrefix(ev.Payload, "swaytab:") {
				switch strings.TrimPrefix(ev.Payload, "swaytab:") {
				case "tab":
					if time.Now().Sub(press) < 600*time.Millisecond {
						currentMode = CYCLE
					} else {
						currentMode = TOGGLE
					}
					var nextWindow int

					switch currentMode {
					case TOGGLE:
						currentWindow = 0
						nextWindow = currentWindow + 1
					case CYCLE:
						nextWindow = currentWindow + 1
						currentWindow = nextWindow
					}

					press = time.Now()

					fmt.Println("tick", currentMode, currentWindow, "->", nextWindow, o)
					focusPos(o, nextWindow)
				}
			}
		}
	}

	err = subscription.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func focusPos(o *org, to int) {
	if len(o.n[o.w[0]]) == 0 {
		return
	}

	to = to % len(o.n[o.w[0]])
	nid := o.n[o.w[0]][to]
	if rc, err := i3.RunCommand(fmt.Sprintf("[con_id=%d] focus", nid)); err != nil {
		log.Println(err, rc)

	}
}
