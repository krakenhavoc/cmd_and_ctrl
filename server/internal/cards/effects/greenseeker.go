package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greenseeker — Creature — Elf Spellshaper {G}, 1/1 (Time Spiral, #5596):
//
//	"{G}, {T}, Discard a card: Search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle."
//
// A Spellshaper (nemSpellshaperCost) whose spell is Bushwhack's first
// bullet: a basic land card, revealed, into the hand, then a shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "693d8203-4405-4488-9067-cc1cd8041cb4",
		Name:         "Greenseeker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}, Discard a card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    nemSpellshaperCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Greenseeker — a basic land card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
