# Site Analyser

A web application that analyses websites and provides detailed information about their structure, including HTML version, headings, links, and login forms.

## Features

- Analyse any website by URL
- Detect HTML version
- Count and categorize headings (h1-h6)
- Identify internal and external links
- Detect login forms
- Check link accessibility

## Prerequisites

- Go 1.21 or higher
- Air (for hot reload during development)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/nuwanwimalasena/site-analyser.git
cd site-analyser
```

2. Install dependencies:
```bash
go mod download
```

3. Install Air for hot reload (optional, for development):
```bash
go install github.com/cosmtrek/air@latest
```

## Development

### Local Development with Hot Reload

1. Start the development server with hot reload:
```bash
air
```

The server will start at `http://localhost:8080` and automatically reload when you make changes to the code.

### Manual Development

If you prefer not to use hot reload, you can run the application directly:

```bash
go run .
```

## Building the Application

### Build for Current Platform

```bash
go build -o build/site-analyser
```

### Build for Specific Platform

```bash
# For Linux
GOOS=linux GOARCH=amd64 go build -o build/site-analyser-linux

# For Windows
GOOS=windows GOARCH=amd64 go build -o build/site-analyser.exe

# For macOS
GOOS=darwin GOARCH=amd64 go build -o build/site-analyser-mac
```

## Running Tests

### Run All Tests

```bash
go test -v .
```

### Run Specific Test File

```bash
go test -v ./ -run . service_test.go
```

### Run Tests with Coverage

```bash
go test -v -cover .
```

## Project Structure

```
site-analyser/
├── internal/
│   ├── controller/
│   │   └── controller.go
│   └── service/
│       └── service.go
├── templates/
│   ├── form.html
│   └── results.html
├── main.go
├── go.mod
└── go.sum
```
## Notes ##

I have decided to build this app using gin web framework, Also I used gin html templates to render the UI instead of using separate FE app. I used bootstrap to make the UI bit nicer. But didn't much focused on FE work since this is more BE related work.

In the BE I think, there are few points we can improve, 
1. We can consider using Goroutines to process the link validations to improve the performance.
2. When Identify a login form, current logic might not accurate, we can consider improvements on there too.
3. Also currently this app analyse the main page only. We can improve this app to read child pages using identified internal links.