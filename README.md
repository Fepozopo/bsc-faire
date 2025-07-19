# bsc-faire

## Graphical User Interface (GUI)

This project now includes a graphical user interface (GUI) built with the [Fyne](https://fyne.io/) framework, making it accessible for non-technical users.

### Features

- **Process Shipments CSV:** Select a CSV file and process shipments, with success/error feedback.
- **Get All Orders:** Enter a sale source (`sm` or `bsc`) to fetch and display all orders.
- **Get Order By ID:** Enter a sale source and order ID to fetch and display a specific order.

### Prerequisites

- Go 1.18 or newer
- [Fyne dependencies](https://developer.fyne.io/started/#prerequisites) (for your OS)
- Set the required environment variables for API tokens:
	- `BSC_API_TOKEN` and/or `SMD_API_TOKEN`

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
*Get all orders by sale source. `sale_source` must be `sm` or `bsc`.*

**Example:**
```
./bsc-faire orders sm
```

#### Get Order By ID

```
./bsc-faire order [sale_source] [orderID]
```
*Get a single order by sale source (`sm` or `bsc`) and order ID.*

**Example:**
```
./bsc-faire order bsc 123456
```

**Note:** Ensure the required environment variables (`BSC_API_TOKEN` and/or `SMD_API_TOKEN`) are set before running commands.

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
2. **Get All Orders:** Click the button, enter `sm` or `bsc` as the sale source, and view the results.
3. **Get Order By ID:** Click the button, enter the sale source and order ID, and view the order details.
