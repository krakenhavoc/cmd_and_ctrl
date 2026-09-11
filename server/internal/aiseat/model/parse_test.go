package model

import "testing"

// parse_test.go pins the reply parser. It is deliberately tolerant of
// packaging — a fenced block or a stray sentence should not throw
// away a correct answer — and strict about content: anything that is
// not ultimately an integer is malformed and the window falls back.

func TestParseAnswer(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
		why  string
		err  bool
	}{
		{name: "the contracted shape", in: `{"index": 3, "why": "removal on the biggest threat"}`, want: 3, why: "removal on the biggest threat"},
		{name: "no why", in: `{"index":0}`, want: 0},
		{name: "index zero", in: `{"index": 0, "why": "pass"}`, want: 0, why: "pass"},
		{name: "whitespace", in: "  \n {\"index\": 7}\n ", want: 7},
		{name: "a bare integer", in: "5", want: 5},
		{name: "a fenced block", in: "```json\n{\"index\": 2, \"why\": \"block\"}\n```", want: 2, why: "block"},
		{name: "leading prose", in: "Sure — here you go:\n{\"index\": 4, \"why\": \"land first\"}", want: 4, why: "land first"},
		{name: "trailing prose", in: `{"index": 1} — holding the bolt for their end step.`, want: 1},
		{name: "a brace inside a string", in: `{"why": "cast {1}{R}", "index": 6}`, want: 6, why: "cast {1}{R}"},
		{name: "an escaped quote inside a string", in: `{"why": "the \"good\" one", "index": 8}`, want: 8, why: `the "good" one`},
		{name: "a negative index parses and is rejected upstream", in: `{"index": -1}`, want: -1},

		{name: "empty", in: "", err: true},
		{name: "whitespace only", in: "   \n\t ", err: true},
		{name: "prose with no number", in: "I would cast the Bear.", err: true},
		{name: "a card name instead of an index", in: `{"index": "Lightning Bolt"}`, err: true},
		{name: "the wrong key", in: `{"move": 3}`, err: true},
		{name: "an unterminated object", in: `{"index": 3`, err: true},
		{name: "a number in prose is not an answer", in: "Move 3 looks strong to me.", err: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, why, err := parseAnswer(tc.in)
			if tc.err {
				if err == nil {
					t.Fatalf("parseAnswer(%q) = %d, want an error", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAnswer(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("index = %d, want %d", got, tc.want)
			}
			if why != tc.why {
				t.Errorf("why = %q, want %q", why, tc.why)
			}
		})
	}
}

func TestFirstJSONObjectIgnoresBracesInStrings(t *testing.T) {
	got := firstJSONObject(`prefix {"a": "{not an object}"} suffix {"b": 1}`)
	if want := `{"a": "{not an object}"}`; got != want {
		t.Errorf("firstJSONObject = %q, want %q", got, want)
	}
	if got := firstJSONObject("no object here"); got != "" {
		t.Errorf("firstJSONObject = %q, want empty", got)
	}
}

func TestModelReasonIsBoundedAndSingleLine(t *testing.T) {
	long := ""
	for i := 0; i < 50; i++ {
		long += "verbose "
	}
	got := modelReason("claude-opus-5", "line one\nline two")
	if got != "claude-opus-5: line one line two" {
		t.Errorf("reason = %q", got)
	}
	if r := modelReason("m", long); len(r) > 130 {
		t.Errorf("reason is %d bytes; a Decision.Reason reaches the chat log", len(r))
	}
	if r := modelReason("m", "   "); r != "m" {
		t.Errorf("reason = %q, want just the model id", r)
	}
}
