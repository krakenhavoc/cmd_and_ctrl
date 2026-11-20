package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrangle — Sorcery {1}{R} (Aether Revolt,
// #5536):
//
//	"Gain control of target creature with power 4 or less until end of
//	 turn. Untap that creature. It gains haste until end of turn."
//
// Act of Treason with a power ceiling on the target, judged at cast and
// again at resolution (threatenAndGrant).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c7d8b6b-9890-4a94-85fc-f7545c4b1cf7",
		Name:         "Wrangle",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with power 4 or less", PowerLE(4)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return threatenAndGrant(ctx, id, "Wrangle", "haste")
		},
	})
}
