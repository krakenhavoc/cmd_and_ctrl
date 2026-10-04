package model

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// openai_schema_test.go pins #2196's half of the OpenAI-compatible
// transport: a move-pick request is constrained to the listed moves.

// A move-pick request carries a json_schema whose index is an enum of
// exactly the listed indices and whose move is an enum of their
// labels, in emission order index → move → why. A server that applies
// it as a grammar cannot write a number that is not listed.
func TestOpenAIConstrainsAMovePickToTheListedMoves(t *testing.T) {
	var raw string
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		raw = string(body)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"index\": 2}"},"finish_reason":"stop"}]}`))
	})
	_, err := c.Complete(context.Background(), Request{
		Model: "m", User: "u",
		Choices: []Choice{{0, "Pass priority"}, {2, "Mountain: Add {R}"}, {3, "Mountain: Add {R}"}, {7, "Cast Bear"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	var got struct {
		ResponseFormat struct {
			Type       string `json:"type"`
			JSONSchema struct {
				Schema struct {
					Properties map[string]struct {
						Type string            `json:"type"`
						Enum []json.RawMessage `json:"enum"`
					} `json:"properties"`
					Required []string `json:"required"`
				} `json:"schema"`
			} `json:"json_schema"`
		} `json:"response_format"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}
	if got.ResponseFormat.Type != "json_schema" {
		t.Fatalf("response_format = %+v, want json_schema:\n%s", got.ResponseFormat, raw)
	}
	props := got.ResponseFormat.JSONSchema.Schema.Properties
	enumText := func(name string) string {
		parts := make([]string, 0, len(props[name].Enum))
		for _, e := range props[name].Enum {
			parts = append(parts, string(e))
		}
		return strings.Join(parts, ",")
	}
	if e := enumText("index"); e != "0,2,3,7" || props["index"].Type != "integer" {
		t.Errorf("index enum = [%s] (%s), want [0,2,3,7] integers — the listed indices and nothing else", e, props["index"].Type)
	}
	if e := enumText("move"); e != `"Pass priority","Mountain: Add {R}","Cast Bear"` {
		t.Errorf("move enum = [%s], want each listed label once", e)
	}
	// Emission order is declaration order on the wire.
	i, m, w := strings.Index(raw, `"index":{`), strings.Index(raw, `"move":{`), strings.Index(raw, `"why":{`)
	if i < 0 || i >= m || m >= w {
		t.Errorf("properties are not in the order index, move, why:\n%s", raw)
	}
	if r := got.ResponseFormat.JSONSchema.Schema.Required; strings.Join(r, ",") != "index,move,why" {
		t.Errorf("required = %v", r)
	}
}

// An improvisation call has no move list, and is never constrained to
// one.
func TestOpenAISendsNoSchemaWithoutChoices(t *testing.T) {
	var raw string
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		raw = string(body)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"},"finish_reason":"stop"}]}`))
	})
	if _, err := c.Complete(context.Background(), Request{Model: "m", User: "u"}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if strings.Contains(raw, "response_format") {
		t.Errorf("a request with no choices carried a response_format:\n%s", raw)
	}
}

// A server that 400s on response_format gets the request once more
// without it, and is not sent the schema again once that succeeds. A
// 400 the retry does not cure is the request's own fault, and the
// schema keeps being sent.
func TestOpenAIDropsTheSchemaForAServerThatRefusesIt(t *testing.T) {
	var calls, withSchema int
	c := serveOpenAI(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "response_format") {
			withSchema++
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"unknown field response_format"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"index\": 0}"},"finish_reason":"stop"}]}`))
	})
	req := Request{Model: "m", User: "u", Choices: []Choice{{0, "Pass priority"}}}
	resp, err := c.Complete(context.Background(), req)
	if err != nil || resp.Text != `{"index": 0}` {
		t.Fatalf("the retry without the schema did not answer: %q, %v", resp.Text, err)
	}
	if calls != 2 || withSchema != 1 || !c.SchemaRefused() {
		t.Fatalf("calls %d, with schema %d, refused %v; want 2, 1, true", calls, withSchema, c.SchemaRefused())
	}
	if _, err := c.Complete(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if calls != 3 || withSchema != 1 {
		t.Errorf("after a refusal: calls %d, with schema %d; want 3, 1 — the schema must stop being sent", calls, withSchema)
	}

	// A 400 the schema did not cause.
	bad := serveOpenAI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"context length exceeded"}}`))
	})
	if _, err := bad.Complete(context.Background(), req); err == nil {
		t.Fatal("a 400 on both attempts returned no error")
	}
	if bad.SchemaRefused() {
		t.Error("a 400 the retry did not cure was blamed on the schema")
	}
}
