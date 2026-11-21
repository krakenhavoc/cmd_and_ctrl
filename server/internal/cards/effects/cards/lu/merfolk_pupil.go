package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Pupil — Creature — Merfolk Wizard {1}{U}, 1/1:
//
//	"When this creature enters, draw a card, then discard a card.
//	 {1}{U}, Exile this card from your graveyard: Draw a card, then discard
//	 a card."
//
// The ETB and the graveyard ability are the same loot (lootOne). The
// activated ability works only from its owner's graveyard and exiles the
// card as a cost (the shape Renew uses), so it can be activated once and
// at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb244878-6440-4dfe-b96b-c9e18d0abe45",
		Name:         "Merfolk Pupil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merfolk Pupil — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Exile this card from your graveyard: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
		}},
	})
}
