# TrackYou - Project Time Tracker

A simple desktop application for tracking time spent on different projects and tasks. Built with Go, Fyne, and SQLite.

## Features

- Track time spent on different projects and tasks
- Start and stop task timers
- View task history with durations
- **Edit past tasks** – modify the project name, description, start time, and end time of any completed task directly from the Log
- **Weekly overview** – per-project totals with daily breakdown (Mon–Sun) for the current calendar week, plus proportional bars
- Persistent storage using SQLite
- Cross-platform support (Windows, macOS, Linux)

## Prerequisites

- Go 1.16 or later
- GCC (for SQLite support)
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/pasmon/trackyou.git
cd trackyou
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build
```

## Usage

1. Run the application:
```bash
./trackyou
```

2. Enter a project name and task description
3. Click "Start Task" to begin timing
4. Click "Stop Task" when finished
5. View your task history in the **Log** tab
6. **Edit a past task**: click the ✏️ (edit) button on any completed task row in the Log to open a dialog where you can update the project name, description, start time, and end time; the duration is recalculated automatically

## Data Storage

All task data is stored locally in a SQLite database file named `tasks.db` located in the user's configuration directory (e.g., `~/.config/TrackYou` on Linux, `~/Library/Application Support/TrackYou` on macOS, `%APPDATA%\TrackYou` on Windows).

## License

MIT License 
