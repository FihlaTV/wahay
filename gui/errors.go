package gui

import (
	"fmt"

	"github.com/coyim/gotk3adapter/gtki"
)

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	//	log.Printf(format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (u *gtkUI) reportError(message string) {
	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", message)
	if err != nil {
		u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
	}

	dlg.SetTransientFor(u.currentWindow)
	u.doInUIThread(func() {
		dlg.Run()
		dlg.Destroy()
	})
}
