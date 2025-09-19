# Faire API Integration & Order Management Tool

## Graphical User Interface (GUI)

This project includes both a command-line interface (CLI) and a graphical user interface (GUI) built with the [Fyne](https://fyne.io/) framework, making it accessible for both technical and non-technical users.

**Now distributed as a single binary that can launch either the CLI or GUI.**

### Features

- **Self-Update:** Both CLI and GUI support self-updating. The GUI checks for updates on startup and via a "Check for Updates" button. Updates are mandatory if a new version is available, and the app will restart after updating.
- **Single Binary:** The project builds a single binary (`faire`) that can launch either the CLI or GUI, depending on how it's started (or via a button in the GUI).
- **Process Shipments CSV:** Select or specify a CSV file and process shipments, with detailed success/error feedback. Failed shipments are shown in the results. The GUI displays a progress bar during shipment processing for a more responsive user experience.
- **Responsive Progress Bar (GUI):** Progress dialogs are shown during all long-running operations, including shipment processing, order fetching, exporting, and self-update.
- **Get All Orders:** Enter a sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) to fetch and display all orders. Internal pagination is always used; CLI flags for limit and page have been removed.
- **Get Order By ID:** Enter a sale source and order ID to fetch and display a specific order.
- **Export NEW Orders to CSV:** Export all new orders for a sale source to a CSV file, including fields like `commission_cents`, `commission_bps`, `item_quantity`, `sale_source`, and more.
- **Test/Mock Mode:** CLI and GUI support mock processing for testing, with options to simulate failures. The GUI now includes a "Use Mock Server" checkbox and a field to specify which shipment indices should fail (e.g., "2,4"). The export command and GUI export now support mock mode, generating a CSV file with only headers (including `sale_source`) and no data.
- **Improved Mock Client:** The mock client now supports both shipment and order operations, including simulating failures for specific shipments, providing mock order data for testing and demos, and generating mock CSV exports.
- **Multi-line Dialogs & Order Formatting:** Order and shipment results are now displayed in improved, multi-line dialogs for better readability. Order formatting in both CLI and GUI has been enhanced.
- **Detailed TUI Feedback:** Both CLI and GUI provide detailed text-based interfaces for viewing processed shipments and orders, including failed shipments.
- **Native File Dialog & Notifications (GUI):** CSV selection uses the system's native file dialog, and results are shown in scrollable dialogs and system notifications.
- **.env Support:** Both CLI and GUI load API tokens and mock settings from a `.env` file if present.
- **Multi-Platform Builds:** Build and run on Windows, Linux, and macOS (ARM/x86_64) using provided Makefile targets. Binaries are named with platform/arch suffixes (e.g., `faire-windows-amd64.exe`).

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

### Building the Project

To build for your platform, use the Makefile targets. This will produce platform-specific binaries in the `bin/` directory (e.g., `faire-windows-amd64.exe`, `faire-linux-amd64`, etc).

```
make windows-amd64   # Windows (AMD64)
make windows-arm64   # Windows (ARM64)
make linux-amd64     # Linux (AMD64)
make linux-arm64     # Linux (ARM64)
make darwin-arm64    # macOS (ARM64)
```

The resulting binary can be used for both CLI and GUI modes.

### Running the CLI

```
./bin/faire --cli
```

### CLI Usage

The CLI provides the following commands (run with `./bin/faire --cli ...`):

#### Process Shipments CSV

```
./bin/faire --cli process [csvfile] [--mock] [--fails indices]
```

_Process shipments from a CSV file and add them to Faire orders._

**Flags:**

- `--mock`: Use a mock Faire client (no real API calls, for testing)
- `--fails`: Comma-separated list of shipment indices to simulate as failures (mock only)

**Example:**

```
./bin/faire --cli process csv/Shipments.csv --mock --fails 2,4
```

_Mock mode improvements: The mock client now supports both shipment and order operations. You can specify which shipment indices should fail using `--fails` (e.g., `--fails 2,4`)._

#### Get All Orders

```
./bin/faire --cli orders [sale_source] [--states STATE1,STATE2]
```

_Get all orders by sale source. `sale_source` can be `21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`._

**Flags:**

- `--states`: Comma-separated list of states to include (if set, only these states are shown)

**Example:**

```
./bin/faire --cli orders bsc --states NEW,PROCESSING
```

#### Get Order By ID

```
./bin/faire --cli order [sale_source] [orderID]
```

_Get a single order by sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) and order ID._

**Example:**

```
./bin/faire --cli order bsc 123456
```

#### Export NEW Orders to CSV

```
./bin/faire --cli export [sale_source] [--mock]
```

_Export all new orders for a sale source to `faire_new_orders.csv`. The CSV includes fields such as `commission_cents`, `commission_bps`, `item_quantity`, `sale_source`, and more._

**Flags:**

- `--mock`: Generate a CSV file with only the headers (including `sale_source`), no real API calls or order data.

**Example:**

```
./bin/faire --cli export bsc --mock
```

**Note:** Ensure the required environment variables (`BSC_API_TOKEN`, `SMD_API_TOKEN`, `C21_API_TOKEN`, `ASC_API_TOKEN`, `BJP_API_TOKEN`, `GTG_API_TOKEN`, `OAT_API_TOKEN`) are set before running commands. For other sale sources, set the corresponding API token environment variable.

### Running the GUI

```
./bin/faire
```

Or double-click the binary to launch the GUI (on supported platforms).

You can also launch the CLI from within the GUI using the "Launch CLI" button.

### GUI Usage

1. **Self-Update:** The GUI checks for updates on startup and via the "Check for Updates" button. Updates are mandatory if available.
2. **Process Shipments CSV:** Click the button and select a CSV file. The app will process the file and show a detailed scrollable dialog and system notification with results, including failed shipments.
3. **Get All Orders:** Click the button, enter any supported sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`), and view the results in a scrollable dialog.
4. **Get Order By ID:** Click the button, enter the sale source and order ID, and view the order details in a scrollable dialog.
5. **Export NEW Orders to CSV:** Click the button, enter the sale source, and export all new orders to `faire_new_orders.csv` (includes commission, item details, and the `sale_source` column). If "Use Mock Server" is checked, the export will generate a CSV file with only the headers (including `sale_source`) and no data.
6. **Mock/Test Mode:** The GUI now uses a "Use Mock Server" checkbox and a field to specify which shipment indices should fail (e.g., "2,4"). The mock mode applies to shipment processing and CSV export.
7. **Improved Dialogs & Formatting:** Order and shipment results are now displayed in improved, multi-line dialogs for better readability. Order formatting in both CLI and GUI has been enhanced.
8. **.env Support:** Both CLI and GUI load API tokens from a `.env` file if present.
9. **Detailed Results:** All dialogs show detailed results, including failed shipments and all shipment/order fields.
10. **Launch CLI:** Use the "Launch CLI" button to open a terminal window running the CLI version of the app.
