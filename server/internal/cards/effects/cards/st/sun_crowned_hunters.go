package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun-Crowned Hunters — Creature — Dinosaur {4}{R}{R}, 5/4 (Ixalan,
// #5532):
//
//	"Enrage — Whenever this creature is dealt damage, it deals 3 damage to
//	 target opponent or planeswalker."
//
// Frilled Deathspitter's shape at 3: one trigger per damage instance
// (CR 120.3), the target picked as it goes on the stack and re-checked as
// it resolves (CR 608.2b). The Hunters are the source even when the
// damage killed them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e6e6458-9e0e-404c-af07-225c4665eaec",
		Name:         "Sun-Crowned Hunters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Cmdr2B08EnrageTrigger("Sun-Crowned Hunters — Enrage: 3 damage to target opponent or planeswalker",
				effects.DealDamageToTheTarget(3)), effects.TargetOpponentOrPlaneswalker()),
		},
	})
}
