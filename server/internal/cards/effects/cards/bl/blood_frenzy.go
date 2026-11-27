package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Frenzy — Instant {1}{R} (Tempest, S82 OLD b755, #5628):
//
//	"Cast this spell only before the combat damage step.
//	 Target attacking or blocking creature gets +4/+0 until end of turn.
//	 Destroy that creature at the beginning of the next end step."
//
// The window is CR 506.8: before the first combat damage step, which in
// practice is a combat before damage, as the target must be attacking or
// blocking. The pump and the delayed destruction are Mogg Cannon's body
// (OldB755TargetGetsThenDestroyAtEndStep), with no keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "8d3d5e87-8b12-4549-a46e-70a103dbcf44",
		Name:               "Blood Frenzy",
		Completeness:       effects.CompletenessFull,
		Targets:            effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:            effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0)),
		CastCondition:      effects.OldB755BeforeTheCombatDamageStep,
		CastConditionLabel: "Cast this spell only before the combat damage step.",
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB755TargetGetsThenDestroyAtEndStep("Blood Frenzy", 4)(ctx.Game, item)
		},
	})
}
