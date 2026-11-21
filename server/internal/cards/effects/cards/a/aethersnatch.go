package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aethersnatch — Instant {4}{U}{U}:
//
//	"Gain control of target spell. You may choose new targets for it.
//	 (If that spell becomes a permanent, it enters under your control.)"
//
// ADR 0104's GainControlOfSpell and nothing else. The steal is a
// layer-2 effect pinned to the spell (CR 611.1, CR 613.1b), so the
// spell's "you" is the thief from then on; the new targets are offered
// to the thief after the change (CR 115.7d); and a permanent spell
// enters under the thief with its caster as its default controller
// (CR 110.2b), which is what sends it home if the thief leaves the
// game (CR 800.4a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45892ec2-8996-444f-a61c-5151926baa0b",
		Name:         "Aethersnatch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.GainControlOfSpell{Spell: t.ID, ChooseNewTargets: true,
				Label: "Aethersnatch — gain control of target spell"}.Apply(ctx)
		},
	})
}
