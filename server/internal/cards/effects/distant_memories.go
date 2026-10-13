package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Distant Memories — Sorcery {2}{U}{U}:
//
//	"Search your library for a card, exile it, then shuffle. Any
//	 opponent may have you put that card into your hand. If no player
//	 does, you draw three cards."
//
// ADR 0150 §3 (#1903): the offer costs nothing (NoPayment). It goes to
// each opponent in APNAP order from the active player, the caster
// skipped, until one says yes. A yes puts the exiled card into the
// caster's hand; if nobody says yes the card stays in exile and the
// caster draws three. With nothing found, a yes still stops the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12c78cc4-7fbe-4639-bef7-bf959a0d6d68",
		Name:         "Distant Memories",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneExile,
				Limit:   1,
				Shuffle: true,
				Reason:  "Distant Memories — search your library for a card to exile",
				Then: func(g *game.Game, found []uuid.UUID) error {
					return distantMemoriesOffer(NewContext(g, item), found)
				},
			}.Apply(ctx)
		},
	})
}

func distantMemoriesOffer(ctx *Context, found []uuid.UUID) error {
	return AnyPlayerPays{
		Pay:          NoPayment(),
		Opponents:    true,
		Question:     "Distant Memories — put the exiled card into its caster's hand, or let them draw three?",
		AcceptLabel:  "Let them have the card",
		DeclineLabel: "Let them draw three",
		OnPaid: func(ctx *Context, _ uuid.UUID) error {
			for _, id := range found {
				if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
					continue
				}
				if err := ctx.Game.BounceToHandForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
		OnNobody: func(ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	}.Apply(ctx)
}
