package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealous Display — Instant {2}{W} (Edge of Eternities, #4461):
//
//	"Creatures you control get +2/+0 until end of turn. If it's not your
//	 turn, untap those creatures."
//
// Borrowed Grace's first mode, whose set of creatures is fixed as the
// spell resolves (CR 611.2c), so a creature that arrives later gets
// nothing. "Those creatures" are that same set: the untap reads the
// creatures its controller controls on the next line of the same
// resolution, with nothing in between that could change who they are. It
// runs only when the caster is not the active player (IsYourTurn), so
// cast on its own turn the spell untaps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ff3afba-7a90-4f62-a1b0-b553ad81c32b",
		Name:         "Zealous Display",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Zealous Display — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			if effects.IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return effects.B16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() })
		},
	})
}
