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

// TestErrorCodes asserts that the error-code constants match the
// exact values specified in docs/protocol.md. Changing a value here is
// a wire-breaking change that must also update docs/protocol.md and the
// TypeScript mirror in client/src/lib/protocol.ts.
func TestErrorCodes(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{CodeBadVersion, "bad_version"},
		{CodeBadJSON, "bad_json"},
		{CodeBadRequest, "bad_request"},
		{CodeInternal, "internal"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("error code: got %q, want %q", c.got, c.want)
		}
	}
}

// TestKindValues asserts the Kind string constants match docs/protocol.md
// verbatim. Same wire-breaking rule as TestErrorCodes.
func TestKindValues(t *testing.T) {
	cases := []struct {
		got  Kind
		want string
	}{
		{KindPing, "ping"},
		{KindPong, "pong"},
		{KindError, "error"},
	}
	for _, c := range cases {
		if string(c.got) != c.want {
			t.Errorf("kind: got %q, want %q", c.got, c.want)
		}
	}
}

func TestVersionIsZero(t *testing.T) {
	if Version != 0 {
		t.Errorf("Version: got %d, want 0 (see docs/protocol.md)", Version)
	}
}
