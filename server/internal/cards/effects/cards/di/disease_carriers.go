package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disease Carriers — Creature — Rat {2}{B}{B}, 2/2 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, target creature gets -2/-2 until end of
//	 turn."
//
// A targeted dies trigger with the shared boostTargetUntilEOT body. The
// target is any player's creature, chosen as the trigger goes on the
// stack and checked again as it resolves (CR 608.2b); the -2/-2 ends in
// the cleanup step (CR 514.2). The target entry declares the harm, so
// the bot aims it at an opponent's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccdd6376-9b74-4d13-9f80-87d0bd3d4a2a",
		Name:         "Disease Carriers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisDies("Disease Carriers — target creature gets -2/-2 until end of turn",
					effects.BoostTargetUntilEOT("Disease Carriers — -2/-2 until end of turn", -2, -2)),
				effects.TargetCreature("target creature"),
			), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
