package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desecrator Hag — Creature — Hag {2}{B/G}{B/G}, 2/2 (Commander 2011, #5574):
//
//	"When this creature enters, return to your hand the creature card in your
//	 graveyard with the greatest power. If two or more cards are tied for
//	 greatest power, you choose one of them."
//
// Not a target: the card is picked as the ability resolves, from the creature
// cards in the controller's graveyard by printed power (a card there has no
// layers; a */* card reads 0). One card with the most power is returned
// without a question; a tie asks the controller to choose among exactly the
// tied cards. An empty graveyard does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "264e60a3-3191-497d-ab06-1411e668b4aa",
		Name:         "Desecrator Hag",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Desecrator Hag — return the creature card with the greatest power from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var tied []uuid.UUID
					best := -1
					for _, c := range p.Graveyard.Cards {
						if !c.IsCreature() {
							continue
						}
						switch {
						case c.Power > best:
							best = c.Power
							tied = []uuid.UUID{c.InstanceID}
						case c.Power == best:
							tied = append(tied, c.InstanceID)
						}
					}
					ctx := effects.NewContext(g, item)
					switch len(tied) {
					case 0:
						return nil
					case 1:
						return effects.ReturnFromGraveyard{Target: tied[0], Dest: game.ZoneHand}.Apply(ctx)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Desecrator Hag — choose a creature card tied for greatest power to return to your hand",
						Cards:    tied,
						Min:      1,
						Max:      1,
						LiveZone: game.ZoneGraveyard,
						ThenKey:  desecratorHagPick,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// desecratorHagPick is the tied card the controller took. An on-disk key.
var desecratorHagPick = effects.AskedPick("choose-cards/desecrator-hag-greatest-power", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})
