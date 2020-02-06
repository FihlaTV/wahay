package gui

import (
	"github.com/coyim/gotk3adapter/gtki"
)

// FileType is the type for identifiying mime-types
type FileType string

var errNoFile = "no file selected"

func (u *gtkUI) setCustomFilePathFor(
	entry gtki.Entry,
	originalValue string,
	onSuccess func(string)) {
	go func() {
		filename := u.getCustomFilePath()

		// The file chooser has been closed or no file has been selected
		if filename == errNoFile {
			return
		}

		if originalValue != filename {
			onSuccess(filename)
			u.doInUIThread(func() {
				entry.SetText(filename)
			})
		}
	}()
}

func (u *gtkUI) getCustomFilePath() string {
	channel := make(chan string)
	go u.showCustomFilePathDialog(channel)
	return <-channel
}

func (u *gtkUI) showCustomFilePathDialog(channel chan string) {
	u.doInUIThread(func() {
		dialog, err := u.g.gtk.FileChooserDialogNewWith2Buttons(
			"Open file",
			u.currentWindow,
			gtki.FILE_CHOOSER_ACTION_OPEN,
			"Cancel",
			gtki.RESPONSE_CANCEL,
			"Open",
			gtki.RESPONSE_ACCEPT)

		if err != nil {
			channel <- errNoFile
			return
		}

		chooser := (dialog).(gtki.FileChooser)
		chooser.SetDoOverwriteConfirmation(true)

		if u.currentWindow != nil {
			dialog.SetTransientFor(u.currentWindow)
		}

		dialog.Present()

		u.disableCurrentWindow()

		res := dialog.Run()

		if gtki.ResponseType(res) == gtki.RESPONSE_ACCEPT {
			channel <- dialog.GetFilename()
		} else {
			channel <- errNoFile
		}

		u.enableCurrentWindow()

		dialog.Destroy()
	})
}
