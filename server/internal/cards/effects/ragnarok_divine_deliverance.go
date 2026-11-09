package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ragnarok, Divine Deliverance — Legendary Creature — Beast Avatar, 7/6:
// the combined back face of Fang, Fearless l'Cie and Vanille, Cheerful
// l'Cie (CR 712.5b); never a card in a deck, only the melded permanent the
// two become (Hanweir, the Writhing Township's shape; ADR 0145):
//
//	"Reach, vigilance, menace, trample, haste
//	 When Ragnarok dies, destroy target permanent and return target
//	 nonlegendary permanent card from your graveyard to the battlefield."
//
// Both targets are chosen as the trigger goes on the stack and each is
// rechecked as it resolves (CR 608.2b): the permanent is destroyed first,
// then the card, which comes back under its owner's control. The melded
// permanent's two cards go to the graveyard together (CR 712.21) but are
// legendary and never a legal second target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "14b839ef-80b1-4e7d-b2fd-38e973899cde",
		Name:            "Ragnarok, Divine Deliverance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "menace", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Ragnarok, Divine Deliverance — destroy target permanent and return target nonlegendary permanent card from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						var ret uuid.UUID
						if len(item.Targets) > 1 && ctx.IsTargetLegal(item.Targets[1]) {
							ret = item.Targets[1].ID
						}
						if len(item.Targets) > 0 && ctx.IsTargetLegal(item.Targets[0]) {
							return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{item.Targets[0].ID},
								withCarry(ctx.ItemTail(ragnarokReturn), ret))
						}
						return ragnarokReturnCard(ctx, ret)
					}),
				TargetPermanent("target permanent").Then(
					TargetCardInGraveyard("target nonlegendary permanent card from your graveyard", YouOwn(), Permanent(), Not(Legendary())))),
		},
	})
}

// ragnarokReturn is "and return target nonlegendary permanent card", run
// once the destroy has finished. The card to return rides the carry. An
// on-disk key.
var ragnarokReturn = CardTail("ragnarok-divine-deliverance/return", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return ragnarokReturnCard(ctx, o.Tail.Carry[0])
})

func ragnarokReturnCard(ctx *Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
}
