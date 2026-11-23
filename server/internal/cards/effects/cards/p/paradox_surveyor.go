package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paradox Surveyor — Creature — Elf Druid {G}{G/U}{U}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Reach
//	 When this creature enters, look at the top five cards of your library.
//	 You may reveal a land card or a card with {X} in its mana cost from
//	 among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order."
//
// Horn of the Mark's sentence (LookAtTopThenMayTakeToHand) with a predicate
// for "a land card or a card with {X} in its mana cost".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "247cc471-fed5-41e3-8b6a-a10bbc129b36",
		Name:            "Paradox Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Paradox Surveyor — look at the top five cards and may take a land card or a card with {X} in its mana cost",
				effects.LookAtTopThenMayTakeToHand(5, effects.SocB31LandOrHasXInManaCost(), 1,
					"Paradox Surveyor — reveal a land card or a card with {X} in its mana cost and put it into your hand")),
		},
	})
}
