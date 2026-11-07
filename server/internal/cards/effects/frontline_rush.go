package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b80d21de-2ace-4e0d-92b2-6f0273512621",
		Name:         "Frontline Rush",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Create two 1/1 red Goblin creature tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: RedGoblinToken(), N: 2}.Apply(ctx)
				}), game.Purpose{Tokens: 2}),
			ModeWithPurpose(ModeDoing("Target creature gets +X/+X until end of turn, where X is the number of creatures you control.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					x := creaturesControlledBy(ctx.Game, ctx.Controller())
					return BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Frontline Rush — +X/+X"}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 0, 0))),
		),
	})
}
