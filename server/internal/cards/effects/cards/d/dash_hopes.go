package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dash Hopes — Instant {B}{B}:
//
//	"When you cast this spell, any player may pay 5 life. If a player
//	 does, counter Dash Hopes.
//	 Counter target spell."
//
// ADR 0150 (#1903): the 5 life is offered to every player in APNAP
// order from the active player, the caster included, until one pays
// (anyPlayerMayCounterThisSpell). A player with less than 5 life, or
// whose life total can't change, can't pay (CR 119.4, 119.8). The
// offer holds Dash Hopes on the stack until it is answered, and the
// first payer counters it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a4ace6b-ef3a-4d11-89d7-958f98dcc445",
		Name:         "Dash Hopes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		Triggered: []game.TriggeredAbility{
			effects.AnyPlayerMayCounterThisSpell("Dash Hopes", "pay 5 life", effects.LifePayment(5), effects.AnyPlayerCounterThisSpell),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
