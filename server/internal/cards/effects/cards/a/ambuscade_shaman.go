package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambuscade Shaman — Creature — Orc Shaman {2}{B}, 2/2 (Dragons of Tarkir, S82
// old batch 331):
//
//	"Whenever this creature or another creature you control enters, that
//	 creature gets +2/+2 until end of turn.
//	 Dash {3}{B} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)"
//
// Bogwater Lumaret's one trigger on every creature entering under your control,
// itself included, with Ogre Battledriver's "that creature" boost: the creature
// that entered gets +2/+2 only while it is still on the battlefield as the
// ability resolves. Tokens count. A dashed Shaman pumps itself as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "96e1d890-4ebb-4c79-8df9-24e3adaf4d9b",
		Name:             "Ambuscade Shaman",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Ambuscade Shaman — that creature gets +2/+2 until end of turn",
				effects.EnteredCreatureGetsPlusTwoPlusTwo("Ambuscade Shaman")),
		},
	})
}
