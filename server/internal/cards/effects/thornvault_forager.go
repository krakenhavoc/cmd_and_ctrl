package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornvault Forager — Creature — Squirrel Ranger {1}{G}, 2/2
// (Bloomburrow, #4467):
//
//	"{T}: Add {G}.
//	 {T}, Forage: Add two mana in any combination of colors. (To forage,
//	 exile three cards from your graveyard or sacrifice a Food.)
//	 {3}{G}, {T}: Search your library for a Squirrel card, reveal it, put
//	 it into your hand, then shuffle."
//
// The second ability is a mana ability (CR 605.1a) whose cost forages
// (#5397, ManaAbilityCost.Forage): the activator names the way — three
// cards from their graveyard or a Food — and with fewer than three cards
// and no Food it can't be activated (CR 118.3). The two mana are two
// independent any-colour slots (AnyCombinationOfColors). The auto-tapper
// never forages on its own; the plain "{T}: Add {G}" is the row it
// plans. "A Squirrel card" is read off the library card's subtypes, so a
// changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b18cfad-a1cf-46c2-bcdd-7762c13d50e5",
		Name:         "Thornvault Forager",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "{T}: Add {G}.",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Forage: true},
				Produced: AnyCombinationOfColors(2),
				Label:    "{T}, Forage: Add two mana in any combination of colors.",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, {T}: Search your library for a Squirrel card, reveal it, put it into your hand, then shuffle.",
			Cost:    Plus(ManaCost("{3}{G}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b06TutorToHand("Thornvault Forager — a Squirrel card, revealed, to hand",
					func(c game.Card) bool { return c.HasSubtypeWord("Squirrel") })(item, NewContext(g, item))
			},
		}},
	})
}
