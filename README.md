# bsc-faire

## Graphical User Interface (GUI)

This project now includes a graphical user interface (GUI) built with the [Fyne](https://fyne.io/) framework, making it accessible for non-technical users.

### Features


- **Process Shipments CSV:** Select a CSV file and process shipments, with detailed success/error feedback. Failed shipments are shown in the results.
- **Get All Orders:** Enter a sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) to fetch and display all orders. Advanced filtering by state, limit, and page is supported in CLI.
- **Get Order By ID:** Enter a sale source and order ID to fetch and display a specific order.
- **Test Process Shipments (GUI):** Preview processed and failed shipments with sample data for demonstration and testing.
- **Detailed TUI Feedback:** Both CLI and GUI provide detailed text-based interfaces for viewing processed shipments and orders, including failed shipments.
- **Native File Dialog (GUI):** CSV selection uses the system's native file dialog for improved usability.
- **Multi-Platform Builds:** Build and run on Windows, Linux, and macOS (ARM/x86_64) using provided Makefile targets.

### Prerequisites

- Go 1.18 or newer (Go 1.23+ recommended)
- [Fyne dependencies](https://developer.fyne.io/started/#prerequisites) (for your OS)
- Set the required environment variables for API tokens:
	- `BSC_API_TOKEN`, `SMD_API_TOKEN`, `C21_API_TOKEN`, `ASC_API_TOKEN`, `BJP_API_TOKEN`, `GTG_API_TOKEN`, `OAT_API_TOKEN` (set as needed for your sale sources)

### Building the CLI

To build the command-line interface (CLI) version:

```
make cli
```

This will produce the `bsc-faire` binary in the project root's `bin/` directory.

### Running the CLI

```
./bsc-faire
```

### CLI Usage

The CLI provides the following commands:

#### Process Shipments CSV

```
./bsc-faire process [csvfile]
```
*Process shipments from a CSV file and add them to Faire orders.*

**Example:**
```
./bsc-faire process csv/Shipments.csv
```

#### Get All Orders

```
./bsc-faire orders [sale_source]
```
*Get all orders by sale source. `sale_source` can be `21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`.*
*Advanced options: use `--limit`, `--page`, and `--states` flags to filter results.*

**Example:**
```
./bsc-faire orders bsc --limit 25 --page 2 --states NEW,PROCESSING
```

#### Get Order By ID

```
./bsc-faire order [sale_source] [orderID]
```
*Get a single order by sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`) and order ID.*

**Example:**
```
./bsc-faire order bsc 123456
```

**Note:** Ensure the required environment variables (`BSC_API_TOKEN` and/or `SMD_API_TOKEN`) are set before running commands.
*For other sale sources, set the corresponding API token environment variable.*

### Building the GUI

```
make gui
```

This will produce the `bsc-faire-gui` binary in the project root's `bin/` directory.

### Running the GUI

```
./bsc-faire-gui
```

### GUI Usage

1. **Process Shipments CSV:** Click the button and select a CSV file. The app will process the file and show a success or error dialog.
2. **Get All Orders:** Click the button, enter any supported sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`), and view the results.
3. **Get Order By ID:** Click the button, enter the sale source and order ID, and view the order details.
4. **Test Process Shipments:** Click the button to preview processed and failed shipments with sample data.
5. **Detailed Results:** All dialogs show detailed results, including failed shipments and all shipment/order fields.
