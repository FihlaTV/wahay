package gui

import (
	"log"

	"autonomia.digital/tonio/app/config"
	"github.com/coyim/gotk3adapter/gtki"
)

type settings struct {
	u      *gtkUI
	b      *uiBuilder
	dialog gtki.Window

	chkAutojoin                gtki.CheckButton
	chkPersistentConfiguration gtki.CheckButton
	chkEncryptFile             gtki.CheckButton
	lblMessage                 gtki.Label
	chkEnableLogging           gtki.CheckButton
	rawLogFile                 gtki.Entry
	btnRawLogFile              gtki.Button
	mumbleBinaryLocation       gtki.Entry

	autoJoinOriginalValue          bool
	persistConfigFileOriginalValue bool
	encryptFileOriginalValue       bool
	logOriginalValue               bool
	rawLogFileOriginalValue        string
	mumbleBinaryOriginalValue      string
}

var (
	decryptUncheckConfirmText = "If you disable this option, anyone could read your configuration settings"
)

func createSettings(u *gtkUI) *settings {
	builder := u.g.uiBuilderFor("GlobalSettings")
	dialog := builder.get("settingsWindow").(gtki.Window)

	s := &settings{
		u:      u,
		b:      builder,
		dialog: dialog,
	}

	s.b.getItems(
		"chkAutojoin", &s.chkAutojoin,
		"chkPersistentConfiguration", &s.chkPersistentConfiguration,
		"chkEncryptFile", &s.chkEncryptFile,
		"lblMessage", &s.lblMessage,
		"chkEnableLogging", &s.chkEnableLogging,
		"rawLogFile", &s.rawLogFile,
		"btnRawLogFile", &s.btnRawLogFile,
		"mumbleBinaryLocation", &s.mumbleBinaryLocation,
	)

	s.init()

	return s
}

func (s *settings) init() {
	conf := s.u.config

	s.autoJoinOriginalValue = conf.GetAutoJoin()
	s.chkAutojoin.SetActive(s.autoJoinOriginalValue)

	s.persistConfigFileOriginalValue = conf.IsPersistentConfiguration()
	s.chkPersistentConfiguration.SetActive(s.persistConfigFileOriginalValue)
	s.lblMessage.SetVisible(!s.persistConfigFileOriginalValue)

	s.encryptFileOriginalValue = conf.ShouldEncrypt()
	s.chkEncryptFile.SetActive(s.encryptFileOriginalValue)
	s.chkEncryptFile.SetSensitive(s.persistConfigFileOriginalValue)

	s.logOriginalValue = conf.IsLogsEnabled()
	s.chkEnableLogging.SetActive(s.logOriginalValue)
	s.rawLogFileOriginalValue = conf.GetRawLogFile()
	s.rawLogFile.SetText(s.rawLogFileOriginalValue)
	s.rawLogFile.SetSensitive(s.logOriginalValue)
	s.btnRawLogFile.SetSensitive(s.logOriginalValue)

	s.mumbleBinaryOriginalValue = conf.GetMumbleBinaryPath()
	s.mumbleBinaryLocation.SetText(s.mumbleBinaryOriginalValue)
}

func (s *settings) processAutojoinOption() {
	conf := s.u.config

	if s.chkAutojoin.GetActive() != s.autoJoinOriginalValue {
		conf.SetAutoJoin(!s.autoJoinOriginalValue)
		s.autoJoinOriginalValue = !s.autoJoinOriginalValue
	}
}

func (s *settings) processPersistentConfigOption() {
	conf := s.u.config

	if s.chkPersistentConfiguration.GetActive() != s.persistConfigFileOriginalValue {
		s.lblMessage.SetVisible(s.persistConfigFileOriginalValue)
		conf.SetPersistentConfiguration(!s.persistConfigFileOriginalValue)
		s.persistConfigFileOriginalValue = !s.persistConfigFileOriginalValue
		s.chkEncryptFile.SetSensitive(s.persistConfigFileOriginalValue)
	}
}

func (s *settings) processEncryptFileOption() {
	conf := s.u.config

	if s.chkEncryptFile.GetActive() != s.encryptFileOriginalValue {
		if s.encryptFileOriginalValue {
			s.u.showConfirmation(func(op bool) {
				if op {
					s.encryptFileOriginalValue = false
					conf.SetShouldEncrypt(false)
					s.chkEncryptFile.SetActive(false)
				} else {
					// We keep the checkbutton checked. Nothing else change.
					s.chkEncryptFile.SetActive(true)
				}
			}, decryptUncheckConfirmText)
		} else {
			s.u.captureMasterPassword(func() {
				s.encryptFileOriginalValue = true
				conf.SetShouldEncrypt(true)
				s.u.saveConfigOnly()
			}, func() {
				s.chkEncryptFile.SetActive(false)
				conf.SetShouldEncrypt(false)
			})
		}
	}
}

func (s *settings) processLogsOption() {
	conf := s.u.config

	if s.chkEnableLogging.GetActive() != s.logOriginalValue {
		s.logOriginalValue = !s.logOriginalValue
		s.rawLogFile.SetSensitive(s.logOriginalValue)
		s.btnRawLogFile.SetSensitive(s.logOriginalValue)
		conf.EnableLogs(s.logOriginalValue)
	}
}

func (u *gtkUI) onSettingsToggleOption(s *settings) {
	s.processAutojoinOption()
	s.processPersistentConfigOption()
	s.processEncryptFileOption()
	s.processLogsOption()
}

func (u *gtkUI) openSettingsWindow() {
	s := createSettings(u)

	cleanup := func() {
		if u.mainWindow != nil {
			u.enableWindow(u.mainWindow)
		}
		s.dialog.Destroy()
		u.currentWindow = nil
	}

	s.b.ConnectSignals(map[string]interface{}{
		"on_toggle_option": func() {
			u.onSettingsToggleOption(s)
		},
		"on_save": func() {
			u.saveConfigOnly()
			cleanup()
		},
		"on_close_window": func() {
			cleanup()
		},
		"on_rawLogFile_icon_press_event":        s.setCustomLogFile,
		"on_rawLogFile_button_clicked_event":    s.setCustomLogFile,
		"on_mumbleBinaryLocation_icon_press":    s.setCustomPathForMumble,
		"on_mumbleBinaryLocation_clicked_event": s.setCustomPathForMumble,
	})

	if u.mainWindow != nil {
		s.dialog.SetTransientFor(u.mainWindow)
		u.disableWindow(u.mainWindow)
	}

	u.currentWindow = s.dialog
	u.doInUIThread(u.currentWindow.Show)
}

func (s *settings) setCustomLogFile() {
	s.u.setCustomFilePathFor(s.rawLogFile, s.rawLogFileOriginalValue, func(f string) {
		s.u.config.SetCustomLogFile(f)
	})
}

func (s *settings) setCustomPathForMumble() {
	s.u.setCustomFilePathFor(s.mumbleBinaryLocation, s.mumbleBinaryOriginalValue, func(f string) {
		s.u.config.SetMumbleBinaryPath(f)
	})
}

func (u *gtkUI) loadConfig() {
	u.config = config.New()

	u.config.WhenLoaded(func(c *config.ApplicationConfig) {
		u.config = c
		u.doInUIThread(u.initialSetupWindow)
		u.configLoaded()
	})

	configFile, err := u.config.DetectPersistence()
	if err != nil {
		log.Fatal("the configuration file can't be initialized")
	}

	if !u.ensureConfig(configFile) {
		u.config.OnAfterLoad()
	} else {
		u.closeApplication()
	}
}

func (u *gtkUI) ensureConfig(configFile string) bool {
	if !u.config.IsPersistentConfiguration() {
		return false
	}

	for {
		isCorrupted, repeatIfFails, err := u.config.LoadFromFile(configFile, u.keySupplier)

		if isCorrupted {
			return u.processCorruptedConfigFileOrExit()
		}

		if repeatIfFails {
			u.keySupplier.Invalidate()
			u.keySupplier.LastAttemptFailed()
			continue
		}

		if err != nil {
			log.Fatal(err)
		}

		break
	}

	return false
}

func (u *gtkUI) processCorruptedConfigFileOrExit() bool {
	if u.regenerateSettingsIfRequiredOrCancel() ||
		u.regenerateEncryptionKeyIfRequiredOrCancel() {
		return true
	}

	u.saveConfigOnly()

	return false
}

func (u *gtkUI) regenerateSettingsIfRequiredOrCancel() bool {
	confirmationChannel := make(chan bool)
	u.askToResetInvalidConfigFile(confirmationChannel)

	if <-confirmationChannel {
		u.config.CreateBackup()
		u.config.DeleteFileIfExists()
		u.config.InitDefault()

		return false
	}

	return true
}

func (u *gtkUI) regenerateEncryptionKeyIfRequiredOrCancel() bool {
	if !u.config.IsFileEncrypted() {
		return false
	}

	passwordChannel := make(chan bool)

	u.captureMasterPassword(func() {
		u.config.SetShouldEncrypt(true)
		passwordChannel <- true
	}, func() {
		u.config.SetShouldEncrypt(false)
		passwordChannel <- false
	})

	selectedOption := <-passwordChannel
	u.config.SetShouldEncrypt(selectedOption)

	return !selectedOption
}

func (u *gtkUI) askToResetInvalidConfigFile(selectionChannel chan bool) {
	u.hideLoadingWindow()

	u.doInUIThread(func() {
		builder := u.g.uiBuilderFor("GlobalSettings")
		dialog := builder.get("winDeleteConfigFileConfirm").(gtki.Window)

		clean := func(op bool) {
			dialog.Destroy()
			u.enableCurrentWindow()
			selectionChannel <- op
		}

		builder.ConnectSignals(map[string]interface{}{
			"on_cancel": func() {
				clean(false)
			},
			"on_delete": func() {
				clean(true)
			},
		})

		dialog.Show()
	})
}

func (u *gtkUI) saveConfigOnlyInternal() error {
	return u.config.Save(u.keySupplier)
}

func (u *gtkUI) saveConfigOnly() {
	// Don't save the configuration file if the user doesn't want it
	if !u.config.IsPersistentConfiguration() {
		u.config.DeleteFileIfExists()
		return
	}

	go func() {
		err := u.saveConfigOnlyInternal()
		if err != nil {
			log.Println("Failed to save config file:", err.Error())
		}
	}()
}
