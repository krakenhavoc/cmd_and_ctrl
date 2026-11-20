package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Shoal — Instant — Arcane {X}{R}{R} (Betrayers of Kamigawa,
// #5602):
//
//	"You may exile a red card with mana value X from your hand rather
//	 than pay this spell's mana cost.
//	 Target creature gets +X/+0 until end of turn."
//
// The pitch is Shining Shoal's (bokB1ShoalPitch): X is announced and the
// exiled red card's mana value must equal it, so pitching a seven-drop
// is +7/+0 for no mana. Hard-cast, X is paid as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ccba143d-ed9b-4044-bdad-f010a0e38e47",
		Name:             "Blazing Shoal",
		XMatters:         true,
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(PowerXUntilEOT(0))},
		Targets:          TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{bokB1ShoalPitch("red", "R")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Blazing Shoal — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}
