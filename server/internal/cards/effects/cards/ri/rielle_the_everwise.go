package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rielle, the Everwise — Legendary Creature — Human Wizard {1}{U}{R}, 0/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Rielle gets +1/+0 for each instant and sorcery card in your graveyard.
//	 Whenever you discard one or more cards for the first time each turn, draw
//	 that many cards."
//
// The first discard of the turn is the first batch of discards, counted from
// the turn's log, so a discard made before Rielle arrived still uses up the
// turn's first; every card of that batch counts, so discarding two draws two.
// A cycled card is a discarded card (CR 702.29d).
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, effects.S82OldB115InstantsAndSorceriesInYourGraveyard)
	pump.Label = "Rielle gets +1/+0 for each instant and sorcery card in your graveyard."
	effects.Register(effects.Spec{
		OracleID:     "6f981e70-5828-49ac-8a72-cf1cc09be991",
		Name:         "Rielle, the Everwise",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventDiscardCard, effects.S82OldB115DiscardedForTheFirstTimeThisTurn,
				"Rielle, the Everwise — draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{N: effects.SettledEventAmount(item)}.Apply(effects.NewContext(g, item))
				}), effects.ThatManyEvents()),
		},
	})
}
