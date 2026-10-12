package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/service/kernel"
	"github.com/hoangnecon/golab/internal/service/notebook"
	"github.com/hoangnecon/golab/internal/service/storage"
	"github.com/hoangnecon/golab/internal/service/telemetry"
)

func printSection(title string, payload any) {
	fmt.Printf("\n======================================================================\n")
	fmt.Printf("📦 TOOL: %s\n", title)
	fmt.Printf("======================================================================\n")
	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting payload: %v\n", err)
		return
	}
	fmt.Println(string(formatted))
}

type ResponseEnvelope struct {
	Status   string           `json:"status"`
	Data     any              `json:"data"`
	Error    *ErrorPayload    `json:"error"`
	Metadata ResponseMetadata `json:"metadata"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponseMetadata struct {
	LatencyMs   int64  `json:"latency_ms"`
	KernelState string `json:"kernel_state,omitempty"`
	Timestamp   string `json:"timestamp"`
}

func main() {
	cdpURL := os.Getenv("COLAB_CDP_URL")
	if cdpURL == "" {
		cdpURL = "http://127.0.0.1:9222"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cdp := bridge.NewCDPClient(cdpURL)
	if err := cdp.Connect(ctx); err != nil {
		log.Fatalf("CDP Connect failed: %v", err)
	}
	defer cdp.Close()

	exec := bridge.NewColabExecutor(cdp)
	storageSvc := storage.NewService(exec)
	kernelSvc := kernel.NewService(exec)
	telemetrySvc := telemetry.NewService(exec)
	notebookSvc := notebook.NewService(exec)

	// 1. Hardware Telemetry
	t0 := time.Now()
	telemData, _ := telemetrySvc.GetHardwareTelemetry(ctx)
	printSection("colab_get_hardware_telemetry", ResponseEnvelope{
		Status: "success",
		Data:   telemData,
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})

	// 2. Drive Status (Disconnected Anti-Hallucination)
	t0 = time.Now()
	driveData, _ := telemetrySvc.CheckDriveStatus(ctx)
	printSection("colab_check_drive_status", ResponseEnvelope{
		Status: "success",
		Data:   driveData,
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})

	// 3. Storage Read Unmounted Drive File (Anti-Hallucination Error Envelope)
	t0 = time.Now()
	_, readErr := storageSvc.ReadFile(ctx, "content/drive/MyDrive/missing_checkpoint.pt", "text")
	printSection("colab_read_file (Drive Unmounted Guardrail)", ResponseEnvelope{
		Status: "error",
		Data:   nil,
		Error: &ErrorPayload{
			Code:    "GOLAB_DRIVE_UNMOUNTED",
			Message: readErr.Error(),
		},
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})

	// 4. Kernel Status
	t0 = time.Now()
	kernelData, _ := kernelSvc.GetStatusReport(ctx)
	printSection("colab_get_kernel_status", ResponseEnvelope{
		Status: "success",
		Data:   kernelData,
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})

	// 5. Notebook Outline
	t0 = time.Now()
	outlineData, _ := notebookSvc.GetNotebookOutline(ctx)
	printSection("colab_get_notebook_outline", ResponseEnvelope{
		Status: "success",
		Data:   outlineData,
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})

	// 6. Storage List Directory
	t0 = time.Now()
	dirData, _ := storageSvc.ListDirectory(ctx, "content")
	printSection("colab_list_directory (/content)", ResponseEnvelope{
		Status: "success",
		Data:   dirData,
		Metadata: ResponseMetadata{
			LatencyMs:   time.Since(t0).Milliseconds(),
			KernelState: "idle",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
	})
}
