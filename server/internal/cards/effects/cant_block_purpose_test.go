package effects

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_block_purpose_test.go — #4977: "can't block this turn" is declared
// as Pump.CantBlock on a harm's target entry, so the bot prices it by the
// attack it opens and not as the removal a harm with no amount (Tragic
// Slip's -X/-X) is priced as.

// declaredHarm is one harm until end of turn the catalog declares on a
// target entry, with the words it was declared for: a card's oracle text,
// a row's label or a mode's.
type declaredHarm struct {
	where, text string
	e           game.TargetPurpose
}

// catalogHarms is every harm until end of turn declared on a target entry
// of a card, an activated row (or its modes) or a spell's mode.
func catalogHarms(t *testing.T) []declaredHarm {
	t.Helper()
	var out []declaredHarm
	add := func(where, text string, p game.Purpose) {
		for _, e := range p.Targets.List() {
			if e.UntilEndOfTurn == game.TargetHarm {
				out = append(out, declaredHarm{where, text, e})
			}
		}
	}
	modes := func(where string, m *game.ModeSpec) {
		if m == nil {
			return
		}
		for i, o := range m.Options {
			add(fmt.Sprintf("%s mode %d %q", where, i, o.Label), o.Label, o.Purpose)
		}
	}
	for _, spec := range All() {
		add(spec.Name, spellText(t, spec.OracleID), spec.Purpose)
		modes(spec.Name, spec.Modes)
		for i, row := range spec.Activated {
			where := fmt.Sprintf("%s ability %d %q", spec.Name, i, row.Label)
			add(where, row.Label, row.Purpose)
			modes(where, row.Modes)
		}
	}
	return out
}

// TestCantBlockDeclarationsMatchTheirText holds every declared CantBlock
// to words that say "can't block", and every harm that declares no amount
// to not leaving one out: a harm with no Pump whose words say "can't
// block" would be priced as removal again. Duel Tactics' "It can't block
// this turn." is a sentence of its own, which the two direction audits'
// pattern does not reach, so this walk reads every declared harm.
func TestCantBlockDeclarationsMatchTheirText(t *testing.T) {
	found := 0
	for _, h := range catalogHarms(t) {
		says := strings.Contains(strings.ToLower(h.text), "can't block")
		pm := h.e.Pump
		switch {
		case pm != nil && pm.CantBlock:
			found++
			if !says {
				t.Errorf("%s declares CantBlock, which its text does not say (#4977)", h.where)
			}
		case pm == nil && says:
			t.Errorf("%s says \"can't block\", declares a harm with no amount, and does not declare CantBlock (#4977)", h.where)
		}
	}
	if found < 4 {
		t.Errorf("found %d CantBlock declarations, want at least the 4 #4977 declared", found)
	}
}

// TestNoAmountHarmsSplitByTheirText: the fifteen harms #4977 found with
// no printed amount, each held to what its text does. The four "can't
// block" cards declare CantBlock and nothing else; the -X/-X family keeps
// the bare direction, which the bot prices as removal.
func TestNoAmountHarmsSplitByTheirText(t *testing.T) {
	cantBlock := map[string]bool{
		"Daring Discovery": true, "Duel Tactics": true, "Fifty Feet of Rope": true, "Untimely Malfunction": true,
	}
	shrinks := map[string]bool{
		"Battle at the Bridge": true, "Call for Blood": true, "Chill Haunting": true, "Defile": true,
		"Grim Hireling": true, "Infused Arrows": true, "Necrotic Wound": true, "Olivia's Midnight Ambush": true,
		"Tragic Slip": true, "Tragic Trajectory": true,
	}
	seen := map[string]bool{}
	for _, h := range catalogHarms(t) {
		name, _, _ := strings.Cut(h.where, " ability ")
		name, _, _ = strings.Cut(name, " mode ")
		pm := h.e.Pump
		switch {
		case cantBlock[name]:
			seen[name] = true
			if pm == nil || !pm.CantBlock || pm.Power != 0 || pm.Toughness != 0 || len(pm.Keywords) > 0 || pm.CantBeBlocked {
				t.Errorf("%s declares %+v, want only CantBlock", h.where, pm)
			}
		case shrinks[name]:
			seen[name] = true
			if pm != nil {
				t.Errorf("%s declares %+v, want the bare direction its -X/-X is priced by", h.where, *pm)
			}
		}
	}
	for name := range cantBlock {
		if !seen[name] {
			t.Errorf("%s declares no harm until end of turn", name)
		}
	}
	for name := range shrinks {
		if !seen[name] {
			t.Errorf("%s declares no harm until end of turn", name)
		}
	}
}

// TestDaringDiscoveryDeclaresCantBlock: the card the issue names, one
// entry for its up-to-three clause and its discover beside it.
func TestDaringDiscoveryDeclaresCantBlock(t *testing.T) {
	s, ok := Lookup("a9564593-b5a6-4a83-be1e-2af8caf647d7")
	if !ok {
		t.Fatalf("Daring Discovery is not registered")
	}
	got := s.Purpose.Targets.List()
	if len(got) != 1 || got[0].UntilEndOfTurn != game.TargetHarm || got[0].Pump == nil || !got[0].Pump.CantBlock {
		t.Errorf("Daring Discovery's target entry = %+v, want a can't-block harm", got)
	}
	if s.Purpose.Discover != 4 {
		t.Errorf("Daring Discovery's discover = %d, want 4", s.Purpose.Discover)
	}
	slip, ok := Lookup("fa6e2a47-a8cb-4224-ae3e-f61c90477285")
	if !ok {
		t.Fatalf("Tragic Slip is not registered")
	}
	if got := slip.Purpose.Targets.List(); len(got) != 1 || got[0].UntilEndOfTurn != game.TargetHarm || got[0].Pump != nil {
		t.Errorf("Tragic Slip's target entry = %+v, want the bare harm", got)
	}
}
