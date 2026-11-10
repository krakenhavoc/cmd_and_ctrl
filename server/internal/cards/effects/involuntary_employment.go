package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7e86fb1e-9987-43ba-b59e-0cb03bbdedb6",
		Name:         "Involuntary Employment",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := threatenAndGrant(ctx, id, "Involuntary Employment", "haste"); err != nil {
				return err
			}
			return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
