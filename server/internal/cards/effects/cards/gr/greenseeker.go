package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "693d8203-4405-4488-9067-cc1cd8041cb4",
		Name:         "Greenseeker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Discard a card: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.NemSpellshaperCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Greenseeker — a basic land card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
