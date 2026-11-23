package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "26c2dd96-07e7-4294-9431-96d5aeff3761",
		Name:         "True Ancestry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Targets:      effects.TargetCardInGraveyard("up to one target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.ClueToken(), N: 1}.Apply(ctx)
		},
	})
}
