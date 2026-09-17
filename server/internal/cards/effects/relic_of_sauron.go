package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relic of Sauron — Artifact {4} (EDHREC rank 1753):
//
//	"{T}: Add two mana in any combination of {U}, {B}, and/or {R}.
//	 {3}, {T}: Draw two cards, then discard a card."
//
// Grixis's four-mana rock. The mana ability is Graven Cairns' shape
// without the filter cost — two pipe slots over the three named
// colours, each answered by its own colour pick, never narrowed
// because the printed text names the colours rather than the
// commander's identity. The loot is an
// ordinary activated ability: the draws land first so a drawn card
// is a legal discard, and the discard is the controller's choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3a81bb1-cbd5-41a0-8dd4-faea06593f84",
		Name:         "Relic of Sauron",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B|R}{U|B|R}",
			Label:    "Add two mana in any combination of {U}, {B}, and/or {R}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Draw two cards, then discard a card.",
			Cost:  Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 2, 1)
			},
		}},
	})
}
