package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frilled Deathspitter — Creature — Dinosaur {2}{R}, 3/2 (Rivals of
// Ixalan, #5531):
//
//	"Enrage — Whenever this creature is dealt damage, it deals 2 damage to
//	 target opponent or planeswalker."
//
// The shared enrage trigger with Final Strike's "target opponent or
// planeswalker" clause, picked as the trigger goes on the stack (CR 603.3d)
// and re-checked as it resolves (CR 608.2b). The Deathspitter is the
// source, even when the damage killed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a65c1fcc-61d1-43d9-a62c-65f156f98bce",
		Name:         "Frilled Deathspitter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Cmdr2B08EnrageTrigger("Frilled Deathspitter — Enrage: 2 damage to target opponent or planeswalker",
				effects.DealDamageToTheTarget(2)), effects.TargetOpponentOrPlaneswalker()),
		},
	})
}
