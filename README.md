# Faire API Integration & Order Management Tool

## Graphical User Interface (GUI)

This project includes both a command-line interface (CLI) and a graphical user interface (GUI) built with the [Fyne](https://fyne.io/) framework, making it accessible for both technical and non-technical users.

### Features

- **Process Shipments CSV:** Select or specify a CSV file and process shipments, with detailed success/error feedback. Failed shipments are shown in the results. The GUI now displays a progress bar during shipment processing for a more responsive user experience.
- **Responsive Progress Bar (GUI):** When processing shipments, the GUI shows a modal progress bar dialog until processing is complete, ensuring users see feedback even for long-running operations.
- **Get All Orders:** Enter a sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) to fetch and display all orders. Advanced filtering by state, limit, and page is supported in CLI.
- **Get Order By ID:** Enter a sale source and order ID to fetch and display a specific order.
- **Test/Mock Mode:** CLI and GUI support mock processing for testing, with options to simulate failures.
- **Detailed TUI Feedback:** Both CLI and GUI provide detailed text-based interfaces for viewing processed shipments and orders, including failed shipments.
- **Native File Dialog & Notifications (GUI):** CSV selection uses the system's native file dialog, and results are shown in scrollable dialogs and system notifications.
- **.env Support (GUI):** The GUI loads API tokens and mock settings from a `.env` file if present.
- **Multi-Platform Builds:** Build and run on Windows, Linux, and macOS (ARM/x86_64) using provided Makefile targets. Binaries are named with platform/arch suffixes (e.g., `faire-cli-linux-x86_64`).

### Developer Notes (Fyne v2.6+)

#### Thread-Safe UI Updates

All Fyne UI updates from goroutines (such as hiding dialogs, showing results, or sending notifications) are performed using `fyne.Do(func() { ... })` as required by Fyne v2.6 and newer. This ensures thread safety and prevents runtime errors. See [Fyne Goroutine Docs](https://docs.fyne.io/started/goroutines.html) for details.

Example:

```
go func() {
	// ...long-running work...
	fyne.Do(func() {
		dialog.ShowCustom(...)
		// ...other UI updates...
	})
}()
```

### Prerequisites

- Go 1.18 or newer (Go 1.23+ recommended)
- [Fyne dependencies](https://developer.fyne.io/started/#prerequisites) (for your OS)
- Set the required environment variables for API tokens:
  - `BSC_API_TOKEN`, `SMD_API_TOKEN`, `C21_API_TOKEN`, `ASC_API_TOKEN`, `BJP_API_TOKEN`, `GTG_API_TOKEN`, `OAT_API_TOKEN` (set as needed for your sale sources)
- (Optional, GUI) Create a `.env` file in the project root to set API tokens and mock settings for the GUI.
  - Example:
    ```
    BSC_API_TOKEN=your_token_here
    FAIRE_USE_MOCK=1
    FAIRE_MOCK_FAILS=2,4
    ```

### Building the CLI

To build the command-line interface (CLI) version:

```
make cli
```

This will produce the `faire-cli-native` binary in the `bin/` directory (or `faire-cli-<platform>` for cross-builds).

### Running the CLI

```
./bin/faire-cli-native
```

### CLI Usage

The CLI provides the following commands:

#### Process Shipments CSV

```
./bin/faire-cli-native process [csvfile] [--mock] [--fails indices]
```

_Process shipments from a CSV file and add them to Faire orders._

**Flags:**

- `--mock`: Use a mock Faire client (no real API calls, for testing)
- `--fails`: Comma-separated list of shipment indices to simulate as failures (mock only)

**Example:**

```
./bin/faire-cli-native process csv/Shipments.csv --mock --fails 2,4
```

#### Get All Orders

```
./bin/faire-cli-native orders [sale_source] [--limit N] [--page N] [--states STATE1,STATE2]
```

_Get all orders by sale source. `sale_source` can be `21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`._

**Flags:**

- `--limit`: Max number of orders to return (10-50, default 50)
- `--page`: Page number to return (default 1)
- `--states`: Comma-separated list of states to include (if set, only these states are shown)

**Example:**

```
./bin/faire-cli-native orders bsc --limit 25 --page 2 --states NEW,PROCESSING
```

#### Get Order By ID

```
./bin/faire-cli-native order [sale_source] [orderID]
```

_Get a single order by sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) and order ID._

**Example:**

```
./bin/faire-cli-native order bsc 123456
```

**Note:** Ensure the required environment variables (`BSC_API_TOKEN`, `SMD_API_TOKEN`, `C21_API_TOKEN`, `ASC_API_TOKEN`, `BJP_API_TOKEN`, `GTG_API_TOKEN`, `OAT_API_TOKEN`) are set before running commands. For other sale sources, set the corresponding API token environment variable.

### Building the GUI

```
make gui
```

This will produce the `faire-gui-native` binary in the `bin/` directory (or `faire-gui-<platform>` for cross-builds).

### Running the GUI

```
./bin/faire-gui-native
```

### GUI Usage

1. **Process Shipments CSV:** Click the button and select a CSV file. The app will process the file and show a detailed scrollable dialog and system notification with results, including failed shipments.
2. **Get All Orders:** Click the button, enter any supported sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`), and view the results in a scrollable dialog.
3. **Get Order By ID:** Click the button, enter the sale source and order ID, and view the order details in a scrollable dialog.
4. **Mock/Test Mode:** If `FAIRE_USE_MOCK=1` is set in your `.env`, the GUI will use a mock client for testing. Use `FAIRE_MOCK_FAILS` to simulate failed shipments (e.g., `FAIRE_MOCK_FAILS=2,4`).
5. **.env Support:** The GUI loads API tokens and mock settings from a `.env` file if present.
6. **Detailed Results:** All dialogs show detailed results, including failed shipments and all shipment/order fields.
