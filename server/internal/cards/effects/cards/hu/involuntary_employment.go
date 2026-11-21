package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Involuntary Employment — Sorcery {3}{R} (Foundations, #4465):
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn. Create a Treasure
//	 token."
//
// Song-Mad Treachery's threatenAndGrant (control, untap, haste, each for
// this turn, in printed order), then the Treasure. A target that is gone
// as the spell resolves fizzles the whole spell, Treasure included (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e86fb1e-9987-43ba-b59e-0cb03bbdedb6",
		Name:         "Involuntary Employment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := effects.ThreatenAndGrant(ctx, id, "Involuntary Employment", "haste"); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
