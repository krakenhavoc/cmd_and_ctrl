package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindred Charge — Sorcery {4}{R}{R}:
//
//	"Choose a creature type. For each creature you control of the
//	 chosen type, create a token that's a copy of that creature. Those
//	 tokens gain haste. Exile them at the beginning of the next end
//	 step."
//
// Chosen as the spell resolves (#2382). The originals are listed once,
// when the answer arrives, so a token copy that is itself of the type
// is not copied again. Haste rides each token's printed keywords, like
// Electroduplicate's; the exile is a CR 603.7 delayed trigger over
// exactly the tokens this spell made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb0b5b5f-54d3-49c6-9ad8-d44a5551acdb",
		Name:         "Kindred Charge",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Kindred Charge — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return kindredChargeCopies(NewContext(g, item), item.Controller, creaturesOfTypeYouControl(g, item.Controller, t))
				})
			return nil
		},
	})
}

func kindredChargeCopies(ctx *Context, controller uuid.UUID, originals []uuid.UUID) error {
	cursor := b25LastEventSeq(ctx.Game)
	for _, id := range originals {
		if err := (CreateTokenCopy{Controller: controller, Copy: id, N: 1, Except: TokenCopyGainsHaste}).Apply(ctx); err != nil {
			return err
		}
	}
	tokens := b27TokensCreatedByAfter(ctx.Game, controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Kindred Charge — exile the tokens",
		Cards: tokens,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
