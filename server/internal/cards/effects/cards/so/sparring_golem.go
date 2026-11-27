package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparring Golem — Artifact Creature — Golem {3}, 2/2 (Invasion, S82 OLD
// #5616):
//
//	"Whenever this creature becomes blocked, it gets +1/+1 until end of turn
//	 for each creature blocking it."
//
// "Becomes blocked" is the once-per-attacker EventBecomesBlocked (CR 509.1h),
// so a double block is one trigger. The blockers are counted as the trigger
// resolves (game.CreaturesBlockingForEffect): one removed from combat in
// response no longer counts, and the bonus is fixed then. A Golem that left
// the battlefield in response gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c54879da-0f1d-4ad0-b082-95f0e4d35685",
		Name:         "Sparring Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Sparring Golem — it gets +1/+1 until end of turn for each creature blocking it",
				effects.OldB652PumpPerBlocker("Sparring Golem — +1/+1 for each creature blocking it")),
		},
	})
}
