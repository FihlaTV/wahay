package gui

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	gmailURL   = "https://mail.google.com/mail/"
	yahooURL   = "http://compose.mail.yahoo.com/"
	outlookURL = "https://dub130.mail.live.com/default.aspx"
)

func (u *gtkUI) switchToMainWindow() {
	u.switchToWindow(u.mainWindow)
}

func (u *gtkUI) switchToWindow(win gtki.ApplicationWindow) {
	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.Show)
}

func (u *gtkUI) copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func (u *gtkUI) messageToLabel(label gtki.Label, message string, seconds int) {
	label.SetVisible(true)
	label.SetText(message)
	time.Sleep(time.Duration(seconds) * time.Second)
	label.SetText("")
	label.SetVisible(false)
}

func (u *gtkUI) enableWindow(win gtki.Window) {
	win.SetSensitive(true)
}

func (u *gtkUI) disableWindow(win gtki.Window) {
	win.SetSensitive(false)
}

func (u *gtkUI) disableCurrentWindow() {
	if u.currentWindow != nil {
		u.disableWindow(u.currentWindow)
	}
}

func (u *gtkUI) enableCurrentWindow() {
	if u.currentWindow != nil {
		u.enableWindow(u.currentWindow)
	}
}
