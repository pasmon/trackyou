# TrackYou - Project Context

## Project Overview
**TrackYou** is a cross-platform desktop application for tracking time spent on projects and tasks. It is built using **Go** and utilizes the **Fyne** toolkit for its graphical user interface. Data persistence is handled by **SQLite**.

### Key Technologies
*   **Language:** Go (1.21+)
*   **GUI Framework:** Fyne (v2)
*   **Database:** SQLite (via `github.com/mattn/go-sqlite3`)
*   **CI/CD:** GitHub Actions & GoReleaser

## Architecture
The project follows a simple structure:
*   `main.go`: Entry point. Contains the `App` struct, UI layout construction, event handlers (start/stop buttons), and theme toggling logic.
*   `models/`: Contains the `Task` struct and related business logic (e.g., `StopTask`, `UpdateDuration`).
*   `database/`: Handles all SQLite interactions, including schema initialization (`InitDB`), and CRUD operations for tasks and preferences.

## Building and Running

### Prerequisites
*   Go 1.21 or later
*   GCC (required for CGO/SQLite)
*   **Linux specific:** X11 development libraries (`libx11-dev`, `xorg-dev`, `libxtst-dev`, `libpng++-dev`)

### Commands
*   **Install Dependencies:** `go mod download`
*   **Run Locally:** `go run main.go`
*   **Build:** `go build -o trackyou`
*   **Test:** `go test -v ./...`
    *   *Note:* CI sets `CGO_ENABLED=1` (required for SQLite) and `FYNE_TEST_SKIP_GUI=1` for tests.

## Development Conventions

*   **Database:** The application uses a local SQLite file (`tasks.db`) stored in the OS-specific user configuration directory. Tables (`tasks`, `preferences`) are created automatically if they don't exist on startup.
*   **UI:** The UI is constructed procedurally in `main.go`. Theme changes are persisted to the database.
*   **Release:** Releases are automated via `goreleaser` (configured in `.goreleaser.yml`), producing binaries for Linux and Windows (amd64/arm64).
*   **Cross-Compilation:** The project uses `fyne-cross` in CI for building Windows binaries from Linux.

## Agent Guidelines

*   **Testing:** Always verify changes by creating and running tests. Use `go test -v ./database/... ./models/...` as a baseline.
*   **Documentation:** When working with frameworks (like Fyne) or tools, explicitly verify the latest documentation using the `context7` MCP server tools (`resolve-library-id` and `query-docs`) to ensure best practices and up-to-date API usage.