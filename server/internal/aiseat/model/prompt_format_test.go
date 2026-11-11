package model

import (
	"strings"
	"testing"
)

// The primer names the game the seat is playing (ADR 0222 slice 1). A
// Commander deck's primer is rulesPrimer, byte for byte, so its cached
// prompt is what it was; a Standard deck's says two players, sixty
// cards, twenty life and no commander.
func TestPrimerNamesTheFormat(t *testing.T) {
	for _, format := range []string{"", "commander", "vintage"} {
		if got := primerFor(format); got != rulesPrimer {
			t.Errorf("primerFor(%q) is not the Commander primer", format)
		}
	}
	for _, sentence := range []string{primerCommanderGame, primerCommanderPlay, primerCommandZone} {
		if strings.Count(rulesPrimer, sentence) != 1 {
			t.Fatalf("rulesPrimer no longer holds %q exactly once; primerFor would not replace it", sentence)
		}
	}
	std := primerFor("standard")
	for _, want := range []string{"two-player game", "Standard, best of one", "20 life", "no commander", "ordinary Standard player"} {
		if !strings.Contains(std, want) {
			t.Errorf("the Standard primer lacks %q", want)
		}
	}
	for _, not := range []string{"four-player", "(Commander)", "Commander player", "command zone"} {
		if strings.Contains(std, not) {
			t.Errorf("the Standard primer still says %q", not)
		}
	}
	// Everything else is the same text: the move-list contract and the
	// answer shape do not depend on the format.
	if !strings.Contains(std, `{"index": <number>, "move":`) || !strings.Contains(std, "You answer with ONE NUMBER") {
		t.Error("the Standard primer lost the answer contract")
	}

	cmd := DeckProfile{Name: "A"}.staticBlocks()
	if cmd[0].Text != rulesPrimer {
		t.Error("a deck with no format does not open on the Commander primer")
	}
	if got := (DeckProfile{Name: "A", Format: "standard"}).staticBlocks(); got[0].Text != std {
		t.Error("a Standard deck does not open on the Standard primer")
	}
}
