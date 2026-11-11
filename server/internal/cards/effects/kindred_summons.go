package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindred Summons — Instant {5}{G}{G}:
//
//	"Choose a creature type. Reveal cards from the top of your library until
//	 you reveal X creature cards of the chosen type, where X is the number of
//	 creatures you control of that type. Put those cards onto the battlefield,
//	 then shuffle the rest of the revealed cards into your library."
func init() {
	Register(Spec{
		OracleID:     "b488e75f-af9c-43e2-8aca-f269fa2c720e",
		Name:         "Kindred Summons",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Kindred Summons — choose a creature type",
				ctx.ItemTail(kindredSummonsTypeChosen))
			return nil
		},
	})
}

var kindredSummonsTypeChosen = creatureTypeChosen("then/kindred-summons-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	player := ctx.Controller()
	x := cmdr2B03CreaturesYouControlOfType(ctx.Game, player, t)
	run, hits := revealUntilCount(ctx, player, func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(t) },
		x, "Kindred Summons — reveal cards until X creature cards of the chosen type")
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			for _, h := range hits {
				if h == c.InstanceID {
					return true
				}
			}
			return false
		},
		All:      true,
		ThenTail: cmdr2B03ShuffleRestIntoLibraryTail,
	}.Apply(ctx)
})
