package testsr

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_mana_test.go — #4980: an activated row that adds mana on
// one numbered resolution of it in a turn declares which one and what
// it adds (Purpose.ResolutionMana), so the bot can price that
// activation by the mana it returns.

// TestSoulbrightSeekerDeclaresItsThirdResolution holds the declaration
// to the printed text and to the count the ability's own body reads.
func TestSoulbrightSeekerDeclaresItsThirdResolution(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.SoulbrightOracle)
	if !ok || len(spec.Activated) != 1 {
		t.Fatal("Soulbright Seeker is not registered with its one activated ability")
	}
	row := spec.Activated[0]
	rm := row.Purpose.ResolutionMana
	if rm == nil || *rm != (game.ResolutionMana{Resolution: 3, Mana: "{R}{R}{R}{R}"}) {
		t.Fatalf("Soulbright Seeker's row declares ResolutionMana %+v, want the third resolution's {R}{R}{R}{R}", rm)
	}
	text := effectstest.SpellText(t, effectstest.SoulbrightOracle)
	if !strings.Contains(text, "third time this ability has resolved this turn, add {R}{R}{R}{R}") {
		t.Errorf("the oracle fixture no longer says what the declaration does: %q", text)
	}
	if row.Label != cardsso.SoulbrightSeekerLabel {
		t.Errorf("the row's label %q is not the label its resolutions are counted under", row.Label)
	}
	var e *game.TargetPurpose
	for _, x := range row.Purpose.Targets.List() {
		if x.Slot == 0 {
			e = &x
		}
	}
	if e == nil || e.UntilEndOfTurn != game.TargetBenefit || e.Pump == nil || len(e.Pump.Keywords) != 1 || e.Pump.Keywords[0] != "trample" {
		t.Errorf("Soulbright Seeker's target entry = %+v, want trample as a benefit until end of turn", e)
	}
}

func TestResolutionManaGuard(t *testing.T) {
	rm := func(n int, mana string) game.Purpose {
		return game.Purpose{ResolutionMana: &game.ResolutionMana{Resolution: n, Mana: mana}}
	}
	cases := []struct {
		name string
		slot effects.PurposeSlot
		p    game.Purpose
		want string // "" = accepted
	}{
		{"an activated row", effects.PurposeOnActivated, rm(3, "{R}{R}{R}{R}"), ""},
		{"two colours", effects.PurposeOnActivated, rm(2, "{G}{U}"), ""},
		{"colourless", effects.PurposeOnActivated, rm(1, "{C}"), ""},
		{"on the card", effects.PurposeOnCard, rm(3, "{R}"), "off an activated row"},
		{"on a triggered row", effects.PurposeOnTriggered, rm(3, "{R}"), "off an activated row"},
		{"on an any-player row", effects.PurposeOnAnyPlayerActivated, rm(3, "{R}"), "off an activated row"},
		{"no resolution", effects.PurposeOnActivated, rm(0, "{R}"), "before the first"},
		{"no mana", effects.PurposeOnActivated, rm(3, ""), "fixed mana symbols"},
		{"generic mana", effects.PurposeOnActivated, rm(3, "{4}"), "fixed mana symbols"},
		{"an X", effects.PurposeOnActivated, rm(3, "{X}"), "fixed mana symbols"},
		{"a hybrid", effects.PurposeOnActivated, rm(3, "{R/G}"), "fixed mana symbols"},
		{"not mana", effects.PurposeOnActivated, rm(3, "four red"), "fixed mana symbols"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				effects.CheckPurpose("Test Card", "ability 0", c.slot, c.p)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}
