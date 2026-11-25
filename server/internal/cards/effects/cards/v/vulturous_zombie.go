package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulturous Zombie — Creature — Plant Zombie {3}{B}{G}, 3/3 (Commander 2016,
// S82 older sets, #5537):
//
//	"Flying
//	 Whenever a card is put into an opponent's graveyard from anywhere, put a
//	 +1/+1 counter on this creature."
//
// One trigger per card (CR 603.2c): a discard, a mill, a death, a countered
// spell and a card put there from exile each count, a token never does
// (CR 108.2b), and a card that was already in a graveyard and moves to
// another does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11a28f60-2456-4ed5-b1cc-03c395f55e59",
		Name:            "Vulturous Zombie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny(append([]game.EventKind{game.EventCounterSpell}, effects.GraveyardArrivalKinds...),
				effects.S82OldB256CardPutIntoAnOpponentsGraveyard,
				"Vulturous Zombie — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
