# Joystick Screen Lock Inhibitor

This is a simple Go program that prevents the screen from locking when a joystick is connected. It continuously monitors the `/dev/input/` directory for joystick devices and uses `dbus-send` to inhibit or uninhibit the screen saver accordingly.

## Features

- Prevents the screen from locking when a joystick is detected.
- Allows the screen to lock when the joystick is disconnected.

## Requirements

- Linux system with D-Bus and `org.freedesktop.ScreenSaver` support.
- Go installed (for build).

## Build

```sh
git clone https://github.com/DublikuntMux/joysticklock.git
cd joysticklock
go build -o joysticklock
```

## Usage

Run the compiled binary:

```sh
./joysticklock
```

To run it in the background:

```sh
nohup ./joysticklock &
```

## How It Works

1. The program scans `/dev/input/` for joystick devices (`js*`).
2. If a joystick is found, it sends a `dbus-send` command to inhibit the screen saver.
3. If the joystick is disconnected, it sends another command to uninhibit the screen saver.
4. The program checks for joystick presence every 5 seconds.

## Troubleshooting

- Ensure `dbus-send` is installed on your system.
- Run the program with `sudo` if it lacks permission to read `/dev/input/`.
- Check if your system uses `org.freedesktop.ScreenSaver`.

## Contributing

Feel free to submit issues or pull requests!

## Author

[DublikuntMax](https://github.com/DublikuntMux)
