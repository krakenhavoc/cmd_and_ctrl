package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reaper from the Abyss — Creature — Demon {3}{B}{B}{B}, 6/6:
//
//	"Flying
//	 Morbid — At the beginning of each end step, if a creature died
//	 this turn, destroy target non-Demon creature."
//
// Deathreap Ritual's trigger with a target: any player's end step
// (b15EndStepBegan), the morbid clause as the intervening if, read off
// the turn tally as the trigger would fire and again as it resolves
// (CR 603.4). The target is picked as the trigger goes on the stack
// (CR 603.3d), and with no non-Demon creature on the battlefield the
// trigger is removed. "Non-Demon" reads effective subtypes, so a
// changeling counts as a Demon and is safe.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3a50d73-3a1f-43aa-ab97-c6c0031bb734",
		Name:            "Reaper from the Abyss",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15EndStepBegan(ev) && effects.B11CreaturesDiedThisTurn(g) > 0
			}, "Reaper from the Abyss — destroy target non-Demon creature (morbid)", reaperFromTheAbyssDestroy),
				effects.TargetCreature("target non-Demon creature", effects.Not(effects.Subtype("Demon")))),
		},
	})
}

// reaperFromTheAbyssDestroy re-checks morbid and destroys the target if
// it is still legal.
func reaperFromTheAbyssDestroy(g *game.Game, item *game.StackItem) error {
	if effects.B11CreaturesDiedThisTurn(g) == 0 {
		return nil
	}
	return effects.DestroyFirstLegalCardTarget(g, item)
}
