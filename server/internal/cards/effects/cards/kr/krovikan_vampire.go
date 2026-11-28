package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Vampire — Creature — Vampire {3}{B}{B}, 3/3 (Ice Age, S82 OLD, #5636):
//
//	"At the beginning of each end step, if a creature dealt damage by this creature this turn died, put
//	 that card onto the battlefield under your control. Sacrifice it when you lose control of this creature."
//
// An intervening "if" (CR 603.4) checked as the end step begins and again on resolution: some creature
// card in a graveyard was put there from the battlefield this turn after the Vampire damaged it (read as
// the objects they were, CR 400.7). Each such card returns under the Vampire's controller's control, and
// each gets its own delayed trigger that sacrifices it when its controller loses control of the Vampire,
// which includes the Vampire leaving the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fa93c01-f9ea-4d48-a040-418d1ca6be19",
		Name:         "Krovikan Vampire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(
				effects.CondBoard{Name: "a creature dealt damage by this creature this turn died", Test: effects.S82OldB843KrovikanVampireHolds},
				effects.On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return source != nil && len(effects.S82OldB843VampireVictims(g, source.InstanceID)) > 0
				}, "Krovikan Vampire — put each creature it damaged that died this turn onto the battlefield under your control", effects.S82OldB843KrovikanVampireEndStep)),
		},
	})
}
