# Joystick Screen Lock Inhibitor

This is a simple Go program that prevents the screen from locking while a joystick is being used. It watches `/dev/input/` for joystick devices, reads their input, and inhibits the screen saver over D-Bus only while there is recent joystick activity.

## Features

- Prevents the screen from locking while joystick input is detected.
- Allows the screen to lock again after a period without joystick input, even if the joystick stays connected.
- Detects joysticks being plugged in and out instantly via file system notifications.

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

Set how long to wait without joystick input before allowing the screen to lock (default `5m`):

```sh
./joysticklock -idle 10m
```

To run it in the background:

```sh
nohup ./joysticklock &
```

## How It Works

1. The program opens existing joystick devices (`js*`) in `/dev/input/` and watches the directory for joysticks being added or removed.
2. It reads events from every joystick. Button presses and axis movements count as input; small axis changes (stick drift) are ignored.
3. On input, it inhibits the screen saver through `org.freedesktop.ScreenSaver`.
4. When no input is seen for the `-idle` duration, it releases the inhibit so the screen can lock normally.

## Troubleshooting

- If a joystick fails to open with a permission error, make sure your user can read `/dev/input/js*` (for example, by being in the `input` group).
- Check if your system uses `org.freedesktop.ScreenSaver`.

## Contributing

Feel free to submit issues or pull requests!

## Author

[DublikuntMax](https://github.com/DublikuntMux)
