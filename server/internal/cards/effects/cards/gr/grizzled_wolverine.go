package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzled Wolverine — Creature — Wolverine {1}{R}{R}, 2/2 (Ice Age, S82 OLD,
// #5636):
//
//	"{R}: This creature gets +2/+0 until end of turn. Activate only during the
//	 declare blockers step, only if at least one creature is blocking this
//	 creature, and only once each turn."
//
// Three gates on one activation: the step, a blocker on this creature, and the
// once-per-turn tally (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +2/+0 until end of turn. Activate only during the declare blockers step, only if at least one creature is blocking this creature, and only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "da4ed552-a85a-47e1-ab67-5c93a9a20048",
		Name:         "Grizzled Wolverine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:      effects.ManaCost("{R}"),
			Condition: effects.AllConditions(effects.OldB835BlockedByAtLeastOneThisDeclareBlockers(), effects.OncePerTurnActivation(label)),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 0, "Grizzled Wolverine — +2/+0 until end of turn"),
		}},
	})
}
