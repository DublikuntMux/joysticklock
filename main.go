package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"
)

const joystickDir = "/dev/input/"

func inhibitScreensaver() (string, error) {
	cmd := exec.Command("dbus-send", "--session", "--print-reply", "--dest=org.freedesktop.ScreenSaver",
		"/org/freedesktop/ScreenSaver", "org.freedesktop.ScreenSaver.Inhibit",
		"string:joystick-prevent-lock", "string:Joystick is active")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`uint32 (\d+)`)
	match := re.FindStringSubmatch(string(output))

	if len(match) > 1 {
		return match[1], nil
	}
	return "", errors.New("no match found")
}

func uninhibitScreensaver(cookie string) error {
	if cookie == "" {
		return nil
	}
	cmd := exec.Command("dbus-send", "--session", "--print-reply", "--dest=org.freedesktop.ScreenSaver",
		"/org/freedesktop/ScreenSaver", "org.freedesktop.ScreenSaver.UnInhibit",
		"uint32:"+cookie)
	return cmd.Run()
}

func joystickConnected() bool {
	files, err := os.ReadDir(joystickDir)
	if err != nil {
		return false
	}
	for _, file := range files {
		if file.Name()[:2] == "js" {
			return true
		}
	}
	return false
}

func main() {
	var cookie string
	var err error

	for {
		if joystickConnected() {
			if cookie == "" {
				fmt.Println("Joystick connected, preventing screen lock...")
				cookie, err = inhibitScreensaver()
				if err != nil {
					fmt.Println("Failed to inhibit screensaver:", err)
					cookie = ""
				}
			}
		} else {
			if cookie != "" {
				fmt.Println("Joystick disconnected, allowing screen lock...")
				if err := uninhibitScreensaver(cookie); err != nil {
					fmt.Println("Failed to uninhibit screensaver:", err)
				}
				cookie = ""
			}
		}
		time.Sleep(5 * time.Second)
	}
}
