package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d64d69bd-1815-4647-90bc-767fad5c7da7",
		Name:         "Haystack",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target creature you control phases out.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PhaseOut{Targets: effects.LegalTargetCards(item, g)}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
