package protocol

import (
	"encoding/json"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	ping := Frame{
		V:    Version,
		Kind: KindPing,
		ID:   "3c2f8d7e-6b5a-4f1e-9a2c-0d8e7f6a5b4c",
	}
	payload, err := json.Marshal(PingPayload{Msg: "hello"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	ping.Payload = payload

	raw, err := json.Marshal(ping)
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}

	var got Frame
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if got.V != Version {
		t.Errorf("v: got %d, want %d", got.V, Version)
	}
	if got.Kind != KindPing {
		t.Errorf("kind: got %q, want %q", got.Kind, KindPing)
	}
	if got.ID != ping.ID {
		t.Errorf("id: got %q, want %q", got.ID, ping.ID)
	}

	var gotPayload PingPayload
	if err := json.Unmarshal(got.Payload, &gotPayload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if gotPayload.Msg != "hello" {
		t.Errorf("msg: got %q, want %q", gotPayload.Msg, "hello")
	}
}

func TestErrorCodes(t *testing.T) {
	for _, code := range []string{CodeBadVersion, CodeBadJSON, CodeBadRequest, CodeInternal} {
		if code == "" {
			t.Errorf("empty error code constant")
		}
	}
}
