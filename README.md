# Crawler-Go

A powerful, enterprise-grade web crawling and video extraction tool built in Go. Crawler-Go specializes in extracting video content from complex web pages that use JavaScript, dynamic loading, obfuscation techniques, and modern web technologies.

## Features

### Core Capabilities
- **Video Detection**: Extracts videos from HTML5 `<video>` tags, embedded players, and custom implementations
- **Multi-Platform Support**: Handles YouTube, Vimeo, custom players, and direct video files
- **Flexible Output Formats**: Supports JSON, CSV, and plain text output

### JavaScript Engine
- **Headless Chrome Integration**: Full browser environment for JavaScript execution
- **Dynamic Content Loading**: Waits for and captures content loaded via AJAX, fetch, or timers
- **User Interaction Simulation**: Mimics human behavior with mouse movements, clicks, and scrolling
- **Network Request Monitoring**: Intercepts and analyzes network requests for video resources

### Advanced Extraction Techniques
- **Base64 Decoding**: Automatically decodes Base64-encoded video URLs
- **WebAssembly Support**: Executes WASM modules to decrypt protected video URLs
- **Token Management**: Refreshes expiring video tokens via API calls
- **Obfuscation Handling**: Penetrates through multiple layers of content protection

### Performance & Reliability
- **Concurrent Processing**: Configurable concurrency for efficient crawling
- **Timeout Management**: Robust timeout handling for JavaScript operations
- **Fallback Mechanisms**: Falls back to static extraction if JavaScript fails
- **Comprehensive Logging**: Detailed logging for debugging and monitoring

## Installation

### Prerequisites
- **Go 1.21+**: [Download Go](https://golang.org/doc/install)
- **Chrome/Chromium 118+**: Required for JavaScript engine functionality

### Quick Start
```bash
# Clone the repository
git clone https://github.com/itxtx/crawler_go.git
cd crawler_go

# Install dependencies
go mod download

# Build the project
make build

# Verify Chrome installation
make verify-chrome
```

### Platform-Specific Chrome Setup

#### macOS
```bash
# Chrome is usually auto-detected at:
/Applications/Google Chrome.app/Contents/MacOS/Google Chrome

# Or set manually:
export CHROME_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
```

#### Linux
```bash
# Install Chrome
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
sudo apt-get update && sudo apt-get install google-chrome-stable

# Set Chrome path if needed
export CHROME_PATH=/usr/bin/google-chrome
```

#### Windows
```bash
# Chrome is usually auto-detected, or set manually:
set CHROME_PATH="C:\Program Files\Google\Chrome\Application\chrome.exe"
```

## Usage

- **CLI Options**:
  ```bash
  ./crawler <base_url> <max_concurrency> <max_pages> [options]
  ```
  - `base_url`: Starting URL for crawling.
  - `max_concurrency`: Maximum concurrent requests.
  - `max_pages`: Maximum pages to scrape.
  - Options include `selectors`, `output_format`, `video_extract`, etc.

## Testing

- **Run Unit Tests**:
  ```bash
  go test ./...
  ```
- **Specific Test Baseline**:
  ```bash
  make test-jsengine
  ```

## Configuration

- Configuration options are specified via command-line arguments or environment variables. Refer to the `config` package for detailed options.


