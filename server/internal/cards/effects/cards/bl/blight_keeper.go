package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight Keeper — Creature — Bat Imp {B}, 1/1 (Jumpstart 2022, #5948):
//
//	"Flying
//	 {7}{B}, {T}, Sacrifice this creature: Target opponent loses 4 life
//	 and you gain 4 life."
//
// Specter of the Fens's drain with an opponent as its target, paid for
// with the sacrifice as a cost (CR 602.2b). The gain happens only when the
// opponent is still a legal target as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1dc73412-c91c-48c9-bde6-b169add79b37",
		Name:            "Blight Keeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{B}, {T}, Sacrifice this creature: Target opponent loses 4 life and you gain 4 life.",
			Cost:    effects.Plus(effects.ManaCost("{7}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TargetOpponentLosesAndYouGain(g, item, 4)
			},
		}},
	})
}
