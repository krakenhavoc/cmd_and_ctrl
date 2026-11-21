package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krav, the Unredeemed — Legendary Creature — Demon {4}{B}, 3/3:
//
//	"Partner with Regna, the Redeemer (When this creature enters,
//	 target player may put Regna into their hand from their library,
//	 then shuffle.)
//	 {B}, Sacrifice X creatures: Target player draws X cards and gains
//	 X life. Put X +1/+1 counters on Krav."
//
// Partner with is the shared CR 702.124j pair: the commander pairing is
// read off the oracle text by internal/deck, and the entry search is
// PartnerWith. The activation is Grim Hireling's variable sacrifice
// (SacrificeX): X is announced with the activation (CR 602.2b), exactly
// X creatures are named and leave together, and the effect reads X back
// with ctx.X(). Krav may be one of the creatures sacrificed; the target
// player still draws and gains, and the counters then have nowhere to
// go. A target player who has left the game (CR 608.2b) gets nothing,
// and Krav still gets his counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "316d49a0-620b-4942-a5e8-04cff15e6c01",
		Name:         "Krav, the Unredeemed",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Krav, the Unredeemed", "Regna, the Redeemer"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice X creatures: Target player draws X cards and gains X life. Put X +1/+1 counters on Krav.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeX("X creatures", effects.Creature())),
			Targets: effects.TargetPlayer("target player"),
			// The life, and then the counters, wait for the prompts before them (a
			// dredge offer, a CR 616 order; CR 608.2c, #4155).
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return kravEachTarget(ctx, func(t game.TargetRef, x int) error {
						return effects.DrawCards{Player: t.ID, N: x}.Apply(ctx)
					})
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return kravEachTarget(ctx, func(t game.TargetRef, x int) error {
						return effects.GainLife{Player: t.ID, Amount: x}.Apply(ctx)
					})
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if x := ctx.X(); x > 0 {
						return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
					}
					return nil
				}),
			),
		}},
	})
}

// kravEachTarget runs do on each legal player target, with X; it does nothing
// when X is 0.
func kravEachTarget(ctx *effects.Context, do func(t game.TargetRef, x int) error) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := do(t, x); err != nil {
			return err
		}
	}
	return nil
}
