package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martial Coup — Sorcery {X}{W}{W} (EDHREC rank 1398):
//
//	"Create X 1/1 white Soldier creature tokens. If X is 5 or more,
//	 destroy all other creatures."
//
// A one-sided Wrath at seven mana and a token maker at less. The
// tokens are created FIRST, then — when X is five or more — every
// creature that is not one of them is destroyed in one simultaneous
// event (DestroyAllMatching, so dies-triggers see the whole batch).
// "Other" is the exclusion by instance ID, so the freshly made
// Soldiers survive and every other creature, the caster's included,
// does not — as printed.
//
// No simplification on the card. Engine gap it shares with every
// wipe in the catalog, reported on #305 rather than fixed here: the
// simultaneous destroy path does not consult indestructible the way
// the single-target path has since #380, so an indestructible
// creature dies to this the same way it dies to Wrath of God.
func init() {
	Register(Spec{
		OracleID:     "2b7c4dab-e432-4b34-b058-3cec5c0d72df",
		Name:         "Martial Coup",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			tokens, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), b11WhiteSoldierToken(), x, game.TokenEntryOptions{})
			if err != nil {
				return err
			}
			if x < 5 {
				return nil
			}
			return DestroyAllMatching{Match: Except(Creature(), b10OneOf(tokens))}.Apply(ctx)
		},
	})
}
