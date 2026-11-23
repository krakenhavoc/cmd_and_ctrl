package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frontline Rush — Instant {R}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one —
//	 • Create two 1/1 red Goblin creature tokens.
//	 • Target creature gets +X/+X until end of turn, where X is the
//	   number of creatures you control."
//
// X is counted as the spell resolves (CR 608.2h), the target included
// when you control it. A target gone by then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b80d21de-2ace-4e0d-92b2-6f0273512621",
		Name:         "Frontline Rush",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Create two 1/1 red Goblin creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.RedGoblinToken(), N: 2}.Apply(ctx)
				}), game.Purpose{Tokens: 2}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +X/+X until end of turn, where X is the number of creatures you control.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					x := effects.CreaturesControlledBy(ctx.Game, ctx.Controller())
					return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Frontline Rush — +X/+X"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		),
	})
}
