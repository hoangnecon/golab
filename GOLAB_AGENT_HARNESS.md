# GOLAB ENTERPRISE AGENT SKILL HARNESS SPECIFICATION

**System Role**: Google Colab Direct Tunnel MCP Controller  
**Protocol Version**: `v2.0.0-harness`  
**Target Platform**: Google Colab Container Environments (via Chrome DevTools Protocol & REST Tunnel Gateway)

---

## 1. Executive Summary & Architecture

GoLab Enterprise provides an agent harness for deep, zero-blocking control over Google Colab instances.

Unlike traditional Jupyter tools that serialize file operations into Python code cells and wait in the kernel queue, GoLab leverages Chrome DevTools Protocol (CDP) to interface directly with the Colab Public Tunnel REST Gateway.

```
[ AI Agent ] ──MCP JSON-RPC──> [ GoLab Server ] ──CDP (WebSocket)──> [ Colab Browser Tab ]
                                                                            │
                                                            (Tunnel Gateway kernelXhr)
                                                                            ▼
                                                             [ Colab Container VM ]
                                                             ├─ colab-fileshim.py (:3453)
                                                             ├─ jupyter-server (:9000)
                                                             ├─ /api/colab/resources
                                                             └─ /_proxy/<port>/
```

---

## 2. Core Operational Directives & Anti-Hallucination Guardrails

Agents operating this harness MUST strictly obey the following operational rules:

### Rule 1: The Google Drive Mount Verification Guardrail
- **Hallucination Trap**: Assuming `/content/drive` or `/content/drive/MyDrive` is automatically accessible.
- **Mandatory Behavior**:
  1. Before attempting to read, write, or list any path inside `content/drive`, the Agent **MUST** first execute `colab_check_drive_status`.
  2. If `mounted: false` (status: `"disconnected"`):
     - **DO NOT** attempt repeated file reads or assume the path does not exist.
     - Explicitly inform the user or execute a code cell containing:
       ```python
       from google.colab import drive
       drive.mount('/content/drive')
       ```
     - The storage engine will actively reject unmounted Drive accesses with an actionable error.

### Rule 2: Out-of-Band Non-Blocking File I/O
- **Anti-Pattern**: Writing Python code snippets (`with open(...)`, `cat > file`) inside code cells to read or write files.
- **Standard Procedure**:
  - Always use `colab_read_file`, `colab_write_file`, `colab_list_directory`, `colab_create_directory`, `colab_delete_file`, and `colab_move_file`.
  - These tools hit `colab-fileshim.py` directly in **20ms–50ms**. They are **100% non-blocking** and work seamlessly even while a deep learning model is actively training in the foreground.

### Rule 3: Kernel Telemetry & Pre-Run VRAM Verification
- **Anti-Pattern**: Launching heavy GPU fine-tuning or large model weights without checking available resources, resulting in CUDA Out of Memory (OOM) or container crash.
- **Standard Procedure**:
  - Execute `colab_get_hardware_telemetry` to verify:
    - `gpus[0].memory_free_mb`: Available Tesla GPU VRAM.
    - `ram.free_gb`: Available host system RAM.
    - `disks[0].used_percent`: Disk capacity.

### Rule 4: Out-of-Band Kernel Interruption (SIGINT)
- **Problem**: A cell runs an infinite loop or stalls during network download, locking the notebook.
- **Standard Procedure**:
  - Execute `colab_interrupt_kernel`.
  - This dispatches an out-of-band `SIGINT` signal directly via Jupyter REST API in **1ms**, stopping execution immediately while preserving all variables and model weights in memory.
  - If execution is unrecoverable, call `colab_restart_kernel` to flush VRAM and reset the environment.

### Rule 5: Precision Notebook DOM Cell Editing
- **Anti-Pattern**: Overwriting an entire 300-line code cell to modify a single variable or function.
- **Standard Procedure**:
  1. Call `colab_get_notebook_outline` to discover cell IDs and structures.
  2. Call `colab_get_cell_with_lines` to inspect line numbers.
  3. Call `colab_edit_cell_lines` with `start_line` and `end_line` (1-based index) to slice and replace exact lines cleanly.

---

## 3. Responses API Envelope Structure

All 25 GoLab MCP tools adhere to the standardized Responses API payload envelope:

```json
{
  "status": "success",
  "data": { ... },
  "error": null,
  "metadata": {
    "latency_ms": 42,
    "kernel_state": "idle",
    "timestamp": "2026-10-12T07:55:00Z"
  }
}
```

When an error occurs:
```json
{
  "status": "error",
  "data": null,
  "error": {
    "code": "GOLAB_EXEC_FAILED",
    "message": "Detailed actionable error message in English."
  },
  "metadata": {
    "latency_ms": 15,
    "kernel_state": "unknown",
    "timestamp": "2026-10-12T07:55:00Z"
  }
}
```

---

## 4. MCP Tool Catalog Quick Reference

### 📂 Storage & Filesystem (Non-blocking via `colab-fileshim.py`)
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_read_file` | `path`, `format` (`text` \| `base64`) | Reads file content in 20-50ms without queue delays. |
| `colab_write_file` | `path`, `content`, `format` | Creates or atomically writes files on VM or Google Drive. |
| `colab_list_directory` | `path` | Lists directory tree with metadata, sizes, and timestamps. |
| `colab_create_directory` | `path` | Creates a new directory path on VM or Google Drive. |
| `colab_delete_file` | `path` | Permanently deletes a file or directory. |
| `colab_move_file` | `old_path`, `new_path` | Moves or renames files and folders instantly. |

### ⚙️ Kernel Lifecycle & Control
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_get_kernel_status` | *(None)* | Real-time execution state (`idle`, `busy`, `starting`), active kernel ID. |
| `colab_interrupt_kernel` | `kernel_id` *(optional)* | Dispatches SIGINT to stop running cell in 1ms without losing RAM state. |
| `colab_restart_kernel` | `kernel_id` *(optional)* | Restarts kernel and flushes GPU VRAM completely. |
| `colab_list_sessions` | *(None)* | Lists all active notebook sessions and attached kernels. |
| `colab_spawn_secondary_kernel` | *(None)* | Spawns a secondary Python 3 kernel for parallel VM jobs. |

### 📊 Hardware & Environment Telemetry
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_get_hardware_telemetry` | *(None)* | Retrieves Tesla GPU VRAM (used/free/total), GPU utilization %, RAM, Disk. |
| `colab_get_build_info` | *(None)* | Returns Google Colab release image tag and build date. |
| `colab_check_drive_status` | *(None)* | Verifies Google Drive mount status (`connected` \| `disconnected`) and instructions. |

### 🖥️ Linux PTY Terminals
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_list_terminals` | *(None)* | Lists active background Linux PTY sessions. |
| `colab_create_terminal` | *(None)* | Launches a new background bash terminal session. |
| `colab_delete_terminal` | `name` | Terminates a background terminal session. |

### 📓 Notebook DOM Precision Editing
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_get_notebook_outline` | *(None)* | Returns skeletal cell outline (indices, types, line counts, headers). |
| `colab_get_cells` | `start_index`, `end_index`, `include_outputs` | Fetches cell contents and execution output blocks. |
| `colab_add_cell` | `index`, `cell_type` (`code` \| `markdown`), `content` | Inserts a new cell at specified position. |
| `colab_edit_cell_lines` | `cell_id`, `start_line`, `end_line`, `new_content` | Precision line slice-and-replace (1-based index). |
| `colab_update_cell` | `cell_id`, `content` | Overwrites entire cell content. |
| `colab_delete_cell` | `cell_id` | Deletes a cell by ID. |
| `colab_move_cell` | `cell_id`, `new_index` | Reorders cell in notebook. |
| `colab_run_code_cell` | `cell_id` | Executes code cell and retrieves outputs. |
| `colab_get_cell_with_lines` | `cell_id` | Fetches code formatted with numbered lines. |

### 🌐 Reverse Proxy Router
| Tool Name | Parameters | Description |
| :--- | :--- | :--- |
| `colab_proxy_forward` | `port`, `subpath`, `method`, `body` | Routes HTTP requests to internal VM services (Gradio :7860, FastAPI :8000, etc.). |
