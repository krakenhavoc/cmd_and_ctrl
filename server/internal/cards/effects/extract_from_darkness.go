package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extract from Darkness — Sorcery {3}{U}{B} (EDHREC rank 3237):
//
//	"Each player mills two cards. Then you put a creature card from
//	 a graveyard onto the battlefield under your control."
//
// A five-mana Reanimate that mills the table first. The mill is
// every live seat, the caster included, two cards each; the
// reanimation is the shared "from A graveyard … under YOUR control"
// body, so a creature taken from an opponent's pile changes hands.
//
// The creature card is chosen AFTER the mill, as the spell resolves
// (CR 608.2c): a card the spell itself milled is a legal pick, and the
// spell can be cast into empty graveyards and simply mill. It is not a
// target, so hexproof-style protections of a graveyard do not apply
// (CR 115.1 / 608.2b: nothing is re-checked but the card still being
// in a graveyard when the choice lands). The pool is every creature card
// in every graveyard, asked in one prompt; the chosen card enters under
// the caster's control, whoever owned it.
func init() {
	Register(Spec{
		OracleID:     "e597d8a1-3bbc-4001-b642-f4421447970f",
		Name:         "Extract from Darkness",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := b02EachPlayerMills(ctx.Game, item, 2); err != nil {
				return err
			}
			controller := item.Controller
			candidates := allGraveyardsCardIDs(ctx, game.Card.IsCreature)
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  controller,
				Source:   item.SourceCardID,
				Question: "Extract from Darkness — put a creature card from a graveyard onto the battlefield under your control",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if len(picked) == 0 {
						return nil
					}
					return ReturnFromGraveyard{Target: picked[0], Dest: game.ZoneBattlefield, Controller: controller}.Apply(NewContext(g, item))
				},
			})
			return nil
		},
	})
}
