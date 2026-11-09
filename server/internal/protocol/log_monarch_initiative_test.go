package protocol

import "testing"

// #4450: a new monarch or initiative holder is a log line; the text names them.
func TestTheLogNarratesTheMonarchAndTheInitiative(t *testing.T) {
	m := LogEvent{Kind: LogMonarch}
	m.actorName = "Pat"
	if got, want := renderLogText(m, "", ""), "Pat becomes the monarch"; got != want {
		t.Errorf("monarch text = %q, want %q", got, want)
	}
	i := LogEvent{Kind: LogInitiative}
	i.actorName = "Pat"
	if got, want := renderLogText(i, "", ""), "Pat takes the initiative"; got != want {
		t.Errorf("initiative text = %q, want %q", got, want)
	}
}
