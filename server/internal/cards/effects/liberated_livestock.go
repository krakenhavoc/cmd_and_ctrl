package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liberated Livestock — Creature — Cat Bird Ox {5}{W}, 4/6 (Wilds of
// Eldraine Commander, #5368):
//
//	"When this creature dies, create a 1/1 white Cat creature token with
//	 lifelink, a 1/1 white Bird creature token with flying, and a 2/4
//	 white Ox creature token. For each of those tokens, you may put an Aura
//	 card from your hand and/or graveyard onto the battlefield attached to
//	 it."
//
// The three tokens are created in the printed order. Then each, in that
// order, offers one choice: any Aura card in my hand or my graveyard, or
// none (the offer is made only when there is a card to put, and a token that
// has already left gets none). The pick goes onto the battlefield attached to
// that token (CR 303.4f: the effect names the host, so the Aura is not asked
// for one); an Aura that can't legally enchant it stays where it was
// (CR 303.4i), and a card already put for an earlier token is gone from the
// pool. The rest of the walk is carried as the list of tokens still to
// offer, as data, so a table waiting on a pick is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d080cb03-b3e5-42e8-b289-2ce2151b4cb8",
		Name:         "Liberated Livestock",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Liberated Livestock — create a Cat, a Bird and an Ox; for each, you may put an Aura card from your hand and/or graveyard onto the battlefield attached to it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var tokens []uuid.UUID
					for _, tmpl := range []game.Card{
						TokenCard("1/1 white Cat with lifelink"),
						TokenCard("1/1 white Bird with flying"),
						TokenCard("2/4 white Ox"),
					} {
						ids, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(tmpl), 1, game.TokenEntryOptions{})
						if err != nil {
							return err
						}
						tokens = append(tokens, ids...)
					}
					return cmdr2B19OfferAuraForNextToken(g, item.Controller, item.SourceCardID, tokens)
				}),
		},
	})
}
