package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirran Mettle — Instant {G} (Mirrodin Besieged, S82 OLD b465, #5576):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Metalcraft — That creature gets +4/+4 until end of turn instead if you
//	 control three or more artifacts."
//
// Metalcraft is an ability word with no rule of its own, so it is read as
// the spell resolves (Galvanic Blast's shape): the artifacts counted are the
// ones its controller has then. A target gone by then fizzles the spell
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35506b6b-6476-4830-9baa-4051c18322f0",
		Name:         "Mirran Mettle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := 2
			if effects.ControlsAtLeast(3, effects.MatchArtifact)(ctx.Game, ctx.Controller(), ctx.Source()) {
				n = 4
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "Mirran Mettle — +N/+N"}.Apply(ctx)
		},
	})
}
