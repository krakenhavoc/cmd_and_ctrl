package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fertile Thicket — Land (Battle for Zendikar):
//
//	"This land enters tapped.
//	 When this land enters, you may look at the top five cards of your
//	 library. If you do, reveal up to one basic land card from among them,
//	 then put that card on top of your library and the rest on the bottom in
//	 any order.
//	 {T}: Add {G}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ba580c9-f933-43d9-b03d-a349faa6c641",
		Name:         "Fertile Thicket",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Fertile Thicket — look at the top five cards; reveal up to one basic land and put it on top, the rest on the bottom",
				effects.S82OldB311LookAtFiveTopBasicLand("Fertile Thicket — reveal up to one basic land card from among them")),
				"Fertile Thicket — look at the top five cards of your library?"),
		},
	})
}
