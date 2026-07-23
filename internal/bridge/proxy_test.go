package bridge

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequestMarshalsParamsAsObject(t *testing.T) {
	params, err := json.Marshal(map[string]any{
		"name": "get_cells",
		"arguments": map[string]any{
			"cellIndexStart": 0,
			"includeOutputs": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  params,
	})
	if err != nil {
		t.Fatal(err)
	}

	var request map[string]any
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	paramsObject, ok := request["params"].(map[string]any)
	if !ok {
		t.Fatalf("params type = %T, want JSON object; payload=%s", request["params"], data)
	}
	if _, ok := paramsObject["arguments"].(map[string]any); !ok {
		t.Fatalf("arguments type = %T, want JSON object; payload=%s", paramsObject["arguments"], data)
	}
}
