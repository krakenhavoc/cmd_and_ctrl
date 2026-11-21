package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskworker — Artifact Creature — Construct {4}, 2/2 (Mirrodin, #5606):
//
//	"Whenever this creature becomes blocked, regenerate it.
//	 {3}: This creature gets +1/+0 until end of turn."
//
// The block trigger watches EventBecomesBlocked, which the engine emits
// once per blocked attacker however many creatures block it (CR 509.3c),
// so one shield goes up per combat. The shield is Albino Troll's
// Regenerate on the trigger's own source (CR 701.19a): the next time
// this turn the Duskworker would be destroyed, it is tapped, its damage
// is removed and it leaves combat instead. A Duskworker that left the
// battlefield before the trigger resolved gets nothing. The pump is the
// ordinary self boost until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d85a83-323b-4e42-88bf-71167a8174e3",
		Name:         "Duskworker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, "Duskworker — regenerate it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Duskworker — +1/+0 until end of turn"),
		}},
	})
}
