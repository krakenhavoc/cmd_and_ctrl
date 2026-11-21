package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:         "ccba143d-ed9b-4044-bdad-f010a0e38e47",
		Name:             "Blazing Shoal",
		XMatters:         true,
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.PowerXUntilEOT(0))},
		Targets:          effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{effects.BokB1ShoalPitch("red", "R")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Blazing Shoal — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}
