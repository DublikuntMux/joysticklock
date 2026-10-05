package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
)

const (
	inputDir = "/dev/input"

	jsEventButton = 0x01
	jsEventAxis   = 0x02
	jsEventInit   = 0x80

	axisThreshold = 4000
)

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
	obj := conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver")
	return obj.Call("org.freedesktop.ScreenSaver.UnInhibit", 0, cookie).Err
}

func isJoystick(path string) bool {
	return strings.HasPrefix(filepath.Base(path), "js")
}

func readDevice(f *os.File, activity chan<- struct{}, closed chan<- *os.File) {
	defer func() { closed <- f }()

	axes := map[uint8]int16{}
	var event [8]byte
	for {
		if _, err := io.ReadFull(f, event[:]); err != nil {
			return
		}

		value := int16(binary.NativeEndian.Uint16(event[4:6]))
		kind := event[6]
		number := event[7]

		if kind&jsEventInit != 0 {
			if kind&^jsEventInit == jsEventAxis {
				axes[number] = value
			}
			continue
		}

		switch kind {
		case jsEventButton:
		case jsEventAxis:
			delta := int(value) - int(axes[number])
			if delta > -axisThreshold && delta < axisThreshold {
				continue
			}
			axes[number] = value
		default:
			continue
		}

		select {
		case activity <- struct{}{}:
		default:
		}
	}
}

func main() {
	idleTimeout := flag.Duration("idle", 5*time.Minute, "allow screen lock after this long without joystick input")
	flag.Parse()

	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalln("Failed to connect to D-Bus:", err)
	}
	defer conn.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed to create file watcher:", err)
	}
	defer watcher.Close()

	if err := watcher.Add(inputDir); err != nil {
		log.Fatalln("Failed to watch", inputDir+":", err)
	}

	activity := make(chan struct{}, 1)
	closed := make(chan *os.File)
	devices := map[string]*os.File{}

	openDevice := func(path string) error {
		if devices[path] != nil || !isJoystick(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		devices[path] = f
		log.Println("Joystick connected:", path)
		go readDevice(f, activity, closed)
		return nil
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalln("Failed to read", inputDir+":", err)
	}
	for _, entry := range entries {
		if err := openDevice(filepath.Join(inputDir, entry.Name())); err != nil {
			log.Println("Failed to open joystick:", err)
		}
	}

	idle := time.NewTimer(*idleTimeout)
	idle.Stop()

	var cookie uint32

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Fatalln("File watcher stopped")
			}

			switch {
			case event.Has(fsnotify.Create):
				if err := openDevice(event.Name); err != nil && !errors.Is(err, os.ErrPermission) {
					log.Println("Failed to open joystick:", err)
				}
			case event.Has(fsnotify.Chmod):
				if err := openDevice(event.Name); err != nil {
					log.Println("Failed to open joystick:", err)
				}
			case event.Has(fsnotify.Remove):
				if f := devices[event.Name]; f != nil {
					f.Close()
					delete(devices, event.Name)
					log.Println("Joystick disconnected:", event.Name)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				log.Fatalln("File watcher stopped")
			}
			log.Println("File watcher error:", err)

		case f := <-closed:
			f.Close()
			if devices[f.Name()] == f {
				delete(devices, f.Name())
				log.Println("Joystick disconnected:", f.Name())
			}

		case <-activity:
			if cookie == 0 {
				cookie, err = inhibitScreensaver(conn)
				if err != nil {
					log.Println("Failed to inhibit screensaver:", err)
				} else {
					log.Println("Joystick input detected, preventing screen lock...")
				}
			}
			idle.Reset(*idleTimeout)

		case <-idle.C:
			if cookie != 0 {
				log.Println("No joystick input, allowing screen lock...")
				if err := uninhibitScreensaver(conn, cookie); err != nil {
					log.Println("Failed to uninhibit screensaver:", err)
				}
				cookie = 0
			}
		}
	}
}
