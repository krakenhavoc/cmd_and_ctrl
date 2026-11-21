package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre Rhymer // Molten Tide — Creature — Elemental Sorcerer {1}{R}{R},
// 3/3 // Instant {R} (preparation card, CR 722):
//
//	"Prowess
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Molten Tide — "Until end of turn, whenever you tap a Mountain for
//	 mana, add an additional {R}."
//
// Molten Tide is a triggered mana ability (CR 605.1b) the spell sets up
// for the turn: a TurnManaTrigger with ControllerOnly, because the card
// says "you" where High Tide and Bubbling Muck say "a player" (#3504).
// It resolves at once and never uses the stack, reads the producing land
// as it was when tapped, and ends at cleanup. An opponent's Mountain
// triggers nothing and hands them nothing.
//
// No simplification.
func init() {
	const id = "8940f503-fc6d-4fae-be74-22c660cae1a7"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Pyre Rhymer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Molten Tide",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TurnManaTrigger{
				Label:          "Molten Tide — add an additional {R}",
				Subtype:        "Mountain",
				Adds:           "{R}",
				ControllerOnly: true,
			}.Apply(ctx)
		},
	})
}
