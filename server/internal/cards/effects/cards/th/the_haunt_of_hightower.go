package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Haunt of Hightower — Legendary Creature — Vampire {4}{B}{B}, 3/3
// (Ravnica Allegiance, #5521):
//
//	"Flying, lifelink
//	 Whenever The Haunt of Hightower attacks, defending player discards a
//	 card.
//	 Whenever a card is put into an opponent's graveyard from anywhere, put a
//	 +1/+1 counter on The Haunt of Hightower."
//
// The second trigger is one per card moved: each card an opponent's
// graveyard gains from any other zone (a discard, a mill, a death, a spell
// resolving) is its own occurrence (CR 603.2c). The player discards by their
// own choice.
//
// No simplification.
// hauntOfHightowerGrow is the second trigger: any card an opponent's graveyard
// gains, from any zone and by any route the engine reports (a move, a discard,
// a mill, a countered spell).
var hauntOfHightowerGrow = game.TriggeredAbility{
	Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
	AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, ok := effects.B06CardInOpponentsGraveyard(ev, source.Controller, g)
		return ok
	},
	Key: "The Haunt of Hightower — put a +1/+1 counter on this creature",
	Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		return game.NewTriggeredItem(source, "The Haunt of Hightower — put a +1/+1 counter on this creature")
	},
	Effect: effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
}

func init() {
	const name = "The Haunt of Hightower"
	effects.Register(effects.Spec{
		OracleID:        "e70f5520-1b9c-4351-8484-30f0dc692e01",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(name+" — defending player discards a card", effects.AttackDefenderDiscardsACard),
			hauntOfHightowerGrow,
		},
	})
}
