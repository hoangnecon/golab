<div align="center">
  <img src="golab.png" width="180" alt="GoLab Logo">
  <h1>GoLab</h1>
  <p><b>Blazingly Fast, Native Golang MCP Server for Google Colab</b></p>

  [![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
  [![Go Version](https://img.shields.io/badge/go-%3E%3D%201.21-00ADD8.svg)](https://golang.org/)
  [![MCP Compatible](https://img.shields.io/badge/MCP-Ready-success.svg)]()
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)]()
</div>

GoLab is a highly optimized, cross-platform server implementing the Model Context Protocol (MCP) for Google Colab. Written completely from scratch in Go, it serves as an independent, native alternative to the Python implementation. Wait no longer—bridge the gap between your local AI IDEs and remote Colab virtual machines instantly.

A major advantage of GoLab over the official Python implementation is its static tool architecture. **GoLab does not require complex client support for `notifications/tools/list_changed`**. This ensures guaranteed out-of-the-box compatibility with **all** standard MCP clients without dropping tools or enforcing strict dynamic listening requirements.

## ✨ Why GoLab?

- **Blazing Fast**: Written in compiled Go, offering near-zero overhead and immediate WebSocket communication.
- **Universal Compatibility**: Works fluently with Cursor, Claude Desktop, Gemini Code Assist, Windsurf, and Antigravity.
- **Cross-Platform**: Zero Python environment headaches. Seamless execution on Windows, macOS, and Linux.
- **Persistent Stability**: Employs static configurations ensuring safe browser reconnection without ephemeral token desyncs.

---

## 🚀 Quick Start

Getting started is identical whether you are running **macOS**, **Linux**, or **Windows**. By utilizing Go's module network, GoLab downloads and builds itself silently in the background.

**Prerequisite:** Ensure [Go (Golang)](https://go.dev/doc/install) is installed on your machine.

Add the following configuration block to your MCP client configuration (e.g., `mcp_config.json` or your Cursor settings):

```json
{
  "mcpServers": {
    "golab": {
      "command": "go",
      "args": ["run", "github.com/hoangnecon/golab/cmd/server@latest"],
      "env": {
        "COLAB_TOKEN": "YOUR-SECURE-TOKEN-HERE",
        "COLAB_WS_PORT": "9090",
        "COLAB_BASE_URL": "colab.research.google.com"
      }
    }
  }
}
```

## ⚙️ Configuration Parameters

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `COLAB_TOKEN` | (Required) A secret passcode of your choosing used to secure the WebSocket connection between your browser and your local agent. | `""` |
| `COLAB_WS_PORT` | (Optional) The local port used by the browser to communicate with GoLab. | `9090` |
| `COLAB_BASE_URL` | (Optional) The Colab domain used when launching new notebook tabs from the IDE. | `colab.research.google.com` |

---

## 📓 Notebook Preparation & Drive Mounting

GoLab drives incredible automation natively, but it must adhere to Google Colab's strict security sandbox boundaries:

1. **Providing Notebook URLs**: When your AI Agent uses the `open_notebook` tool, it needs the direct Google Drive Colab File URL (e.g., `https://colab.research.google.com/drive/1YUds...`). GoLab will intercept this and automatically append the necessary MCP WebSocket mapping parameters.
2. **Mounting Google Drive**: If your workflow requires the AI agent to read/write files directly to Google Drive via the `list_drive` or `read_file` tools, **the user must manually mount Drive** inside the Colab UI first. You can do this in two ways:
   * **Method A (Native UI):** Click the "Folder" icon in the left sidebar of the Colab interface, then click the "Mount Drive" icon.
   * **Method B (Python Cell):** Execute the standard Drive mounting code:
     ```python
     from google.colab import drive
     drive.mount('/content/drive')
     ```
   *Note: Google strictly enforces an interactive OAuth consent popup for mounting Drive. The GoLab Agent cannot click "Accept" for you. Once mounted, the Agent has full native filesystem control.*

---

## 🧰 Available Tools Matrix

GoLab exposes 15 powerful tools to grant your AI comprehensive control over remote execution pipelines.

| Tool | Category | Description |
|------|--------|-------------|
| **`check_status`** | Connection | Pings the WebSocket to verify if the frontend proxy is actively attached. |
| **`open_notebook`** | Setup | Opens a specified Colab or Drive URL, automatically injecting MCP parameters. |
| **`get_notebook_outline`**| Read | Returns the full skeletal structure of the notebook (cell types, previews). |
| **`get_cells`** | Read | Fetches the raw string block contents of specific cells. |
| **`add_code_cell`** | Write | Appends new Python execution cells natively. |
| **`add_text_cell`** | Write | Appends Markdown notation cells sequentially. |
| **`update_cell`** | Write | Complete overwrite mechanism for an existing structural cell. |
| **`delete_cell`** | Write | Purges target cells from the active Virtual Machine. |
| **`insert_in_cell`** | Editor | Precision injection of raw syntax at specific index lines. |
| **`edit_cell_lines`** | Editor | Slice-and-replace localized lines directly avoiding cell corruption. |
| **`find_replace_in_cell`**| Editor | Perform string-to-string substitution natively inside a cell. |
| **`run_code_cell`** | Execution | Executes the target cell directly on Colab's hardware (e.g., L4 GPU/TPU). |
| **`get_cell_output`** | Read | Retrieves executed `Stdout`, Tracebacks, and Base64 stream data output. |
| **`get_error_cells`** | Debug | Sweeps the active notebook isolating kernel compilation or runtime failures. |
| **`read_file`** / `list_drive` | File I/O | Extracts datasets directly off Google Drive or the local VM `/content/` path. |

---

## 📚 Concrete API Examples

Instead of prompting in natural language, here is exactly how your AI Agent and GoLab communicate payload schemas back-and-forth under the hood using the MCP Protocol:

### 1. Executing a Cell & Reading Output natively
**Agent Payload (`run_code_cell`)**:
```json
{
  "cellId": "jH5vSMErxJwI"
}
```
**GoLab Response**:
```json
{
  "outputs": [
    {
      "output_type": "stream",
      "name": "stdout",
      "text": ["Epoch 1/50 - loss: 0.4213\n"]
    }
  ]
}
```

### 2. Slicing Model Architecture Parameters Directly
**Agent Payload (`insert_in_cell`)**:
```json
{
  "cellId": "g24HDD5s9MNh",
  "lineNumber": 4,
  "content": "import torch\nimport torch.nn as nn"
}
```
**GoLab Response**:
```json
{
  "inserted": true,
  "linesInserted": 2,
  "totalLines": 15
}
```

---

## 🤝 Contributing & Issues

We welcome community contributions to make GoLab even better! 

1. Check the [Issues](https://github.com/hoangnecon/golab/issues) tab for existing reports or feature requests.
2. Fork the repository and create your feature branch.
3. Ensure your Go code is formatted (`go fmt`) and passes standard checks (`go vet`).
4. Submit a Pull Request with a clear description of your architectural changes.

---

## 📜 License

This project is an independent native Golang implementation. It is open-sourced under the [Apache License 2.0](LICENSE).
