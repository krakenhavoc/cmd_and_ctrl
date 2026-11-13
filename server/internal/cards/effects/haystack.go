package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haystack — Artifact {1}{W} (Assassin's Creed, #5466):
//
//	"{2}, {T}: Target creature you control phases out. (Treat it and
//	 anything attached to it as though they don't exist until your
//	 next turn.)"
//
// Vodalian Illusionist's ability, limited to your own creatures. The
// engine does the rest of phasing (docs/adding-cards.md, "Phasing"):
// what is attached goes with it, and it phases in during your next
// untap step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d64d69bd-1815-4647-90bc-767fad5c7da7",
		Name:         "Haystack",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Target creature you control phases out.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PhaseOut{Targets: legalTargetCards(item, g)}.Apply(NewContext(g, item))
			},
		}},
	})
}
