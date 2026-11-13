package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Swords — Sorcery {1}{B}
// (Nemesis, #5618):
//
//	"If you control a Swamp, you may sacrifice a creature rather than pay this spell's mana cost.
//	 Each player exiles two cards from their hand."
//
// nemSacrificeACreatureWithASwamp, then each player, the caster
// included, chooses two cards in their own hand and exiles them (all of
// them with fewer than two). The choices are made in seat order,
// starting with the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4af20f8c-2ba0-49d5-a64e-d3668cec30cd",
		Name:             "Mind Swords",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{nemSacrificeACreatureWithASwamp()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range tablePlayers(ctx) {
				p := ctx.Game.PlayerByIDForEffect(id)
				if p == nil || p.Eliminated || p.Hand == nil || p.Hand.Size() == 0 {
					continue
				}
				hand := make([]uuid.UUID, 0, p.Hand.Size())
				for _, c := range p.Hand.Cards {
					hand = append(hand, c.InstanceID)
				}
				n := min(2, len(hand))
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  id,
					Source:   item.SourceCardID,
					Question: "Mind Swords — exile two cards from your hand",
					Cards:    hand,
					Min:      n,
					Max:      n,
					LiveZone: game.ZoneHand,
					ThenKey:  mindSwordsExile,
				})
			}
			return nil
		},
	})
}

// mindSwordsExile exiles the cards each player chose from their hand. An
// on-disk key: never renamed, never reused.
var mindSwordsExile = game.RegisterChooseCardsThen("choose-cards/mind-swords-exile-two", func(g *game.Game, r game.CardsChosen) error {
	g.ExileCardsForEffect(r.Picked)
	return nil
})
