package main

import (
	"fmt"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
)

const joystickDir = "/dev/input/"

func inhibitScreensaver(conn *dbus.Conn) (uint32, error) {
	obj := conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver")

	var cookie uint32
	err := obj.Call("org.freedesktop.ScreenSaver.Inhibit", 0, "joystick-prevent-lock", "Joystick is active").Store(&cookie)
	if err != nil {
		return 0, err
	}

	return cookie, nil
}

func uninhibitScreensaver(conn *dbus.Conn, cookie uint32) error {
	if cookie == 0 {
		return nil
	}

	obj := conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver")
	return obj.Call("org.freedesktop.ScreenSaver.UnInhibit", 0, cookie).Err
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
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Println("Failed to connect to D-Bus:", err)
		return
	}
	defer conn.Close()

	var cookie uint32

	for {
		if joystickConnected() {
			if cookie == 0 {
				fmt.Println("Joystick connected, preventing screen lock...")
				cookie, err = inhibitScreensaver(conn)
				if err != nil {
					fmt.Println("Failed to inhibit screensaver:", err)
					cookie = 0
				}
			}
		} else {
			if cookie != 0 {
				fmt.Println("Joystick disconnected, allowing screen lock...")
				if err := uninhibitScreensaver(conn, cookie); err != nil {
					fmt.Println("Failed to uninhibit screensaver:", err)
				}
				cookie = 0
			}
		}
		time.Sleep(5 * time.Second)
	}
}
