package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ethereal Valkyrie — Creature — Spirit Angel {4}{W}{U}, 4/4 (Kaldheim
// Commander, #5506):
//
//	"Flying
//	 Whenever this creature enters or attacks, draw a card, then exile a
//	 card from your hand face down. It becomes foretold. Its foretell cost
//	 is its mana cost reduced by {2}. (On a later turn, you may cast it for
//	 its foretell cost, even if this creature has left the battlefield.)"
//
// One trigger for both events. The controller draws, then CHOOSES the card
// (a keyed prompt; nothing is asked of an empty hand). It goes to exile face
// down as a foretold card through Game.ForetellCardFromHandForEffect: its
// foretell cost is its mana cost with {2} of generic mana taken off (never
// into a coloured symbol), it may be cast on a later turn, and the permission
// belongs to the card, not to the Valkyrie. It is an effect, not the foretell
// special action, so nothing is paid and "the first card you foretell each
// turn" does not see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f0f7c3df-52f8-4835-a18f-69e56ffa3432",
		Name:            "Ethereal Valkyrie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Ethereal Valkyrie — draw a card, then exile a card from your hand face down; it becomes foretold",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
						return nil
					}
					cards := make([]uuid.UUID, 0, p.Hand.Size())
					for _, c := range p.Hand.Cards {
						cards = append(cards, c.InstanceID)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:    item.Controller,
						FromPlayer: item.Controller,
						Source:     item.SourceCardID,
						Question:   "Ethereal Valkyrie — exile a card from your hand face down; it becomes foretold",
						Cards:      cards,
						Min:        1,
						Max:        1,
						ThenKey:    etherealValkyrieForetell,
						LiveZone:   game.ZoneHand,
					})
					return nil
				}),
		},
	})
}

// etherealValkyrieForetell is the keyed answer: foretell the pick, its cost
// two generic less. An on-disk key: never renamed, never reused.
var etherealValkyrieForetell = game.RegisterChooseCardsThen("choose-cards/ethereal-valkyrie-foretell", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := g.ForetellCardFromHandForEffect(r.Chooser, id, 2); err != nil {
			return err
		}
	}
	return nil
})
