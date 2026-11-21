package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kessig Flamebreather — Creature — Human Shaman {1}{R}, 1/3 (EDHREC
// rank 1464):
//
//	"Whenever you cast a noncreature spell, this creature deals 1
//	 damage to each opponent."
//
// Firebrand Archer with a bigger backside. The condition is the
// shared "noncreature spell cast by you" reader — the spell's type
// line is read off the stack — and the damage is the creature's, so
// a doubler doubles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eadd9559-6dcb-4b96-8c95-58abddd0e120",
		Name:         "Kessig Flamebreather",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Kessig Flamebreather — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
