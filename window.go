package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
)

func gtkActivated(app *gtk.Application) {

	settings, _ := gtk.SettingsGetDefault()
	settings.Set("gtk-application-prefer-dark-theme", true)

	css, err := gtk.CssProviderNew()
	if err != nil {
		log.Fatal("Unable to create css provider:", err)
	}
	err = css.LoadFromData(`
* { box-shadow: none; border-radius: 0; }
.window-frame { box-shadow: none; }
.window-frame:backdrop { box-shadow: none; }
.header-bar { box-shadow: none; }
.window-frame.csd.popup { box-shadow: none; }
.window-frame.solid.csd { box-shadow: none; }
decoration { box-shadow: none; }
decoration:backdrop { box-shadow: none; }
`)
	if err != nil {
		log.Fatal("Unable to load css:", err)
	}

	win, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	screen := win.GetScreen()
	gtk.AddProviderForScreen(screen, css, 600)

	header, err := gtk.DialogNew()
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitlebar(header)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.Connect("key-press-event", func(window *gtk.ApplicationWindow, event *gdk.Event, app *gtk.Application) {
		onKeypress(app, event)
	}, app)

	// Create a new label widget to show in the window.
	l, err := gtk.LabelNew("Hello, gotk3!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	// Add the label to the window.
	win.Add(l)

	// Set the default window size.
	win.SetDefaultSize(800, 600)

	// Recursively show all widgets contained in this window.
	win.ShowAll()
}

func onKeypress(app *gtk.Application, event *gdk.Event) {
	ke := gdk.EventKeyNewFromEvent(event)
	if ke.State()&gdk.GDK_CONTROL_MASK != 0 && ke.KeyVal() == gdk.KEY_q {
		app.Quit()
	} else if ke.KeyVal() == gdk.KEY_h {
		win := app.GetActiveWindow()
		if a, err := win.GetProperty("visible"); err == nil {
			if a.(bool) == true {
				win.Hide()
				glib.TimeoutAdd(uint(5000), func() { glib.IdleAdd(win.Show) })
			}
		}
	}
}
