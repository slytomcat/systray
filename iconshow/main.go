package main

import (
	"fmt"
	"time"

	"github.com/slytomcat/systray"
)

var (
	iconNames = []string{
		"network-error",
		"network-idle",
		"network-offline",
		"network-receive",
		"network-transmit",
		"security-high",
		"security-medium",
		"security-low",
	}
)

func main() {
	onExit := func() {
		now := time.Now()
		fmt.Println("Exit at", now.String())
	}

	systray.Run(onReady, onExit)
}

func setTitle(name string) {
	systray.SetTitle(fmt.Sprintf("icon show: %s", name))
}

func setMenuTitle(m *systray.MenuItem, name string) {
	m.SetTitle(fmt.Sprintf("Change Icon to %s", name))

}

func onReady() {
	systray.SetIconByName(iconNames[0])
	iconID := 1
	setTitle(iconNames[0])
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()

	// We can manipulate the systray in other goroutines
	go func() {

		mChange := systray.AddMenuItem("Change Icon", "Change Icon")
		setMenuTitle(mChange, iconNames[1])
		for range mChange.ClickedCh {
			systray.SetIconByName(iconNames[iconID])
			setTitle(iconNames[iconID])
			iconID = (iconID + 1) % len(iconNames)
			setMenuTitle(mChange, iconNames[iconID])
		}
	}()
	mQuit.Enable()
}
