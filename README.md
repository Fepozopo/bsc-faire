# Faire API Integration & Order Management GUI

A graphical application for managing Faire orders and shipments. It is built with [Fyne](https://fyne.io/) and contains no command-line interface.

## Features

- **Process shipments CSV:** Select a CSV file and add its shipments to Faire orders, with detailed success and failure feedback.
- **Get all orders:** Fetch and display orders for a supported sale source (`21`, `asc`, `bjp`, `bsc`, `gtg`, `oat`, or `sm`).
- **Get order by ID:** Retrieve and display one order by its sale source and display ID.
- **Export NEW orders:** Export new orders for a sale source to `faire_new_orders.csv`, including commission, item quantity, and sale source fields.
- **Mock/test mode:** Use the mock client for demos and tests, including optional simulated shipment failures.
- **Self-update:** Check for application updates at startup or with the **Check for Updates** button.
- **Native file selection and notifications:** Use the system file picker to choose CSV files and display operation results in the GUI.
- **`.env` support:** Load API tokens and mock settings from an optional `.env` file.
- **Multi-platform builds:** Build binaries for Windows, Linux, and macOS using the supplied Makefile targets.

## Prerequisites

- Go 1.23 or newer
- [Fyne dependencies](https://developer.fyne.io/started/#prerequisites) for your operating system
- API tokens for the sale sources you use:
  `BSC_API_TOKEN`, `SMD_API_TOKEN`, `C21_API_TOKEN`, `ASC_API_TOKEN`, `BJP_API_TOKEN`, `GTG_API_TOKEN`, and `OAT_API_TOKEN`

Optionally, create a `.env` file in the project root:

```dotenv
BSC_API_TOKEN=your_token_here
FAIRE_USE_MOCK=1
FAIRE_MOCK_FAILS=2,4
```

## Building

Use a Makefile target to build the GUI binary for a supported platform:

```sh
make windows-amd64
make windows-arm64
make linux-amd64
make linux-arm64
make darwin-arm64
```

The resulting platform-specific binary is created in `bin/`.

## Running

Run the platform-specific GUI binary directly. For example, on Apple Silicon macOS:

```sh
./bin/faire-darwin-arm64
```

You can also double-click the binary on platforms that support it.

## GUI usage

1. **Process Shipments CSV:** Select a CSV file, confirm it, and view the detailed result dialog.
2. **Get All Orders:** Enter a supported sale source to retrieve its active orders.
3. **Get Order By ID:** Enter the sale source and display ID to view one order.
4. **Export NEW Orders to CSV:** Enter a sale source to create `faire_new_orders.csv`.
5. **Mock/Test Mode:** Enable **Use Mock Server** and optionally specify failing shipment indices such as `2,4`.
6. **Check for Updates:** Use the button to manually check for a newer application version.
