package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immobilizer Eldrazi — {1}{R} Creature — Eldrazi Drone 2/1 (Oath of the
// Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 {2}{C}: Each creature with toughness greater than its power can't block
//	 this turn."
//
// Goblin Locksmith's query-filtered can't-block scope: the set is every
// creature whose toughness is greater than its power, read live until end of
// turn (CR 611.2c), so a creature that becomes one later is covered and one
// that stops being one is freed. The {C} in the cost is paid with colorless
// mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "416f9656-9247-4ef2-971b-1515fc8e031b",
		Name:            "Immobilizer Eldrazi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{C}: Each creature with toughness greater than its power can't block this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{C}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RestrictUntilEOT{
					Scope:        game.ScopeCreaturesMatching,
					Restrictions: game.CantBlock,
					Queries:      []game.PermanentQuery{{Types: []string{"Creature"}, ToughnessOverPower: true}},
					Label:        "Immobilizer Eldrazi — creatures with toughness greater than power can't block",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
