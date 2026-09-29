package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Regal Behemoth — Creature — Dinosaur {4}{G}{G}, 5/5:
//
//	"Trample
//	 When this creature enters, you become the monarch.
//	 Whenever you tap a land for mana while you're the monarch, add an
//	 additional one mana of any color."
//
// The green monarch fatty (#1722). The mana line is a CR 605.1b
// TRIGGERED MANA ability (ADR 0074): it resolves the instant the land's
// ability has, with no stack and no window, so the extra mana is there
// for the spell it is paying for. "While you're the monarch" is read as
// the land is tapped — YoureTheMonarch against the Behemoth's
// controller — so the bonus stops the moment an opponent takes the
// crown and comes back when you take it back, with no bookkeeping.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "797c2b1c-c373-4735-b397-f561559c7c61",
		Name:            "Regal Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Regal Behemoth"),
		},
		ManaTriggers: []game.ManaTrigger{{
			Label: "Regal Behemoth — add one mana of any color",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
				return prod.Source.IsLand() && prod.Controller == source.Controller &&
					YoureTheMonarch(g, source.Controller)
			},
			Produced: AddsFixedMana("{W|U|B|R|G}"),
		}},
	})
}
