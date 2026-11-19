package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bb244878-6440-4dfe-b96b-c9e18d0abe45",
		Name:         "Merfolk Pupil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Merfolk Pupil — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) }),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, Exile this card from your graveyard: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{U}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
		}},
	})
}
