package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Under the Skin — Sorcery {2}{G}:
//
//	"Manifest dread. (Look at the top two cards of your library. Put
//	 one onto the battlefield face down as a 2/2 creature and the other
//	 into your graveyard. Turn it face up any time for its mana cost if
//	 it's a creature card.)
//	 You may return a permanent card from your graveyard to your hand."
//
// The second sentence is not a target, so it is asked as a pick when
// the manifest has finished: the card manifest dread just put into the
// graveyard is a candidate, which is the point of the card. "You may"
// is a pick with a floor of zero; with no permanent card there to take,
// nobody is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e121178-e9e3-43a5-a174-ecbe47d759f9",
		Name:         "Under the Skin",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me, source := ctx.Controller(), ctx.Source()
			return ManifestDread{Then: func(g *game.Game, _ game.ManifestDreadResult) error {
				return underTheSkinOffer(g, me, source)
			}}.Apply(ctx)
		},
	})
}

// underTheSkinOffer asks `me` which permanent card in their graveyard
// to take, if any, and returns it to their hand.
func underTheSkinOffer(g *game.Game, me, source uuid.UUID) error {
	candidates := graveyardCardIDs(NewContext(g, nil), me, game.Card.IsPermanent)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   source,
		Question: "Under the Skin — you may return a permanent card from your graveyard to your hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.ReturnFromGraveyardUnderControlForEffect(id, game.ZoneHand, uuid.Nil); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
