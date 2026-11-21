package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assassin Den — Creature — Wall {3}{U}, 2/4 (Assassin's Creed, #5466):
//
//	"Defender (This creature can't attack.)
//	 {3}{U}: Put a +1/+1 counter on target creature you control. It
//	 can't be blocked this turn. Activate only as a sorcery."
//
// Two steps on one target: the counter, then the restriction, so the
// second waits for any prompt the counter asks (a CR 616 order).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "826274fd-9831-4122-a5d9-0d1caa3d3a85",
		Name:            "Assassin Den",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U}: Put a +1/+1 counter on target creature you control. It can't be blocked this turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:         effects.ManaCost("{3}{U}"),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			SorcerySpeed: true,
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.TargetCantBeBlockedThisTurn("Assassin Den")(ctx.Game, ctx.Item)
				}),
			),
		}},
	})
}
