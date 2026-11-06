package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// True Ancestry — Sorcery — Lesson {1}{G} (Avatar: The Last Airbender,
// #4459):
//
//	"Return up to one target permanent card from your graveyard to your
//	 hand.
//	 Create a Clue token. (It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Walk with the Ancestors' return, then a Clue. With no target chosen the
// spell still resolves and makes the Clue; with a target chosen that is
// gone as it resolves, the spell has no legal target and does nothing
// (CR 608.2b). Creating the Clue is not investigating (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26c2dd96-07e7-4294-9431-96d5aeff3761",
		Name:         "True Ancestry",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Targets:      TargetCardInGraveyard("up to one target permanent card from your graveyard", YouOwn(), Permanent()).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Controller: ctx.Controller(), Template: ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
