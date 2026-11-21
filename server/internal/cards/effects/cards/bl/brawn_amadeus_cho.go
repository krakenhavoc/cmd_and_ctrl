package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brawn, Amadeus Cho — {1}{G/U} Legendary Creature — Gamma Scientist
// Hero 1/1 (Marvel Super Heroes Commander, #4455):
//
//	"When Brawn enters, draw a card.
//	 Power-up — {4}{G/U}: Put a +1/+1 counter on Brawn for each card in
//	 your hand. (Activate each power-up ability only once. Reduce the
//	 cost by his mana cost if he entered this turn.)"
//
// Brave Brawler's power-up row with the counter count read from the
// controller's hand as the ability resolves, so the draw from the enters
// trigger and any card played in response are counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19910a90-d334-4d99-89f2-cd81cea3c824",
		Name:         "Brawn, Amadeus Cho",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Brawn — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.PowerUpAnswering("{4}{G/U}: Put a +1/+1 counter on Brawn for each card in your hand.",
				effects.ManaCost("{4}{G/U}"), brawnCountersForCardsInHand),
		},
	})
}

// brawnCountersForCardsInHand is "put a +1/+1 counter on this creature
// for each card in your hand".
func brawnCountersForCardsInHand(g *game.Game, item *game.StackItem) error {
	n := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		n = p.Hand.Size()
	}
	return effects.PlusOneCountersOnThis(n)(g, item)
}
