package main

import (
	"fmt"
	"go.i3wm.org/i3/v4"
)

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

func (o *org) WindowFrontID(nid i3.NodeID) {
	for _, w := range o.m {
		for _, n := range w {
			if n.ID == nid {
				o.WindowFront(n)
			}
		}
	}
}
