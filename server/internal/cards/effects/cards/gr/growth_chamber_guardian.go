package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Growth-Chamber Guardian — Creature — Elf Crab Warrior {1}{G}, 2/2 (Ravnica
// Allegiance, #5521):
//
//	"{2}{G}: Adapt 2. (If this creature has no +1/+1 counters on it, put two
//	 +1/+1 counters on it.)
//	 Whenever one or more +1/+1 counters are put on this creature, you may
//	 search your library for a card named Growth-Chamber Guardian, reveal it,
//	 put it into your hand, then shuffle."
//
// One trigger per counter placement (CR 603.2c). The search is for a stated
// quality, so the library is shuffled whether or not the player finds a card
// (CR 701.23b).
//
// No simplification.
func init() {
	const name = "Growth-Chamber Guardian"
	effects.Register(effects.Spec{
		OracleID:     "1a49ec1c-6150-4b46-b671-b7d6f4e301ae",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCounterPlaced, effects.GapB51CountersPlacedOnThis,
				name+" — you may search your library for a card named "+name,
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, name) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Source:    item.SourceCardID,
						Reason:    name + " — search for a card named " + name,
					}.Apply(effects.NewContext(g, item))
				}),
				name+" — search for a card named "+name+"?"),
		},
	})
}
