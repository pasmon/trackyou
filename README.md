# TrackYou - Project Time Tracker

A simple desktop application for tracking time spent on different projects and tasks. Built with Go, Fyne, and SQLite.

## Features

- Track time spent on different projects and tasks
- Start and stop task timers
- View task history with durations
- Persistent storage using SQLite
- Cross-platform support (Windows, macOS, Linux)

## Prerequisites

- Go 1.16 or later
- GCC (for SQLite support)
- Git

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/trackyou.git
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
5. View your task history in the list below

## Data Storage

All task data is stored locally in a SQLite database file named `tasks.db` in the application directory.

## License

MIT License 