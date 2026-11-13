package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind's Desire — Sorcery {4}{U}{U} (Scourge):
//
//	"Shuffle your library. Then exile the top card of your library. Until
//	 end of turn, you may play that card without paying its mana cost.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)"
//
// Urza, Lord High Artificer's {5} ability as a spell: shuffle, exile the
// top card, and grant its controller a free play of it until end of turn
// (a land is played under the usual land rules, a spell is cast for {0}),
// plus the storm keyword (CR 702.40a). Each copy shuffles and exiles its
// own card, and each exiled card carries its own permission.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3739b5a-5731-4c64-a244-815a363b0d5c",
		Name:         "Mind's Desire",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			if err := (ShuffleLibrary{Player: me}).Apply(ctx); err != nil {
				return err
			}
			_, err := ctx.Game.ExileTopWithPermissionForEffect(me, me, 1, game.CastPermission{
				Cost:     "{0}",
				Duration: ctx.Game.UntilEndOfTurnDuration(),
			})
			return err
		},
	})
}
