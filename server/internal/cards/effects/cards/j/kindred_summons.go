package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindred Summons — Instant {5}{G}{G}:
//
//	"Choose a creature type. Reveal cards from the top of your library until
//	 you reveal X creature cards of the chosen type, where X is the number of
//	 creatures you control of that type. Put those cards onto the battlefield,
//	 then shuffle the rest of the revealed cards into your library."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b488e75f-af9c-43e2-8aca-f269fa2c720e",
		Name:         "Kindred Summons",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Kindred Summons — choose a creature type",
				ctx.ItemTail(kindredSummonsTypeChosen))
			return nil
		},
	})
}

var kindredSummonsTypeChosen = effects.CreatureTypeChosen("then/kindred-summons-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	player := ctx.Controller()
	x := effects.Cmdr2B03CreaturesYouControlOfType(ctx.Game, player, t)
	run, hits := effects.RevealUntilCount(ctx, player, func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(t) },
		x, "Kindred Summons — reveal cards until X creature cards of the chosen type")
	return effects.PutFromLibraryOntoBattlefield{
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
		ThenTail: effects.Cmdr2B03ShuffleRestIntoLibraryTail,
	}.Apply(ctx)
})
