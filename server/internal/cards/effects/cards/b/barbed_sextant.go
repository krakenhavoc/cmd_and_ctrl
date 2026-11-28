package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barbed Sextant — Artifact {1} (Ice Age, S82 OLD batch 837, #5636):
//
//	"{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card at
//	 the beginning of the next turn's upkeep."
//
// Chromatic Star's mana ability. It has no target, adds mana and moves no card to
// or from a library (the draw is a delayed trigger, not part of the ability), so
// it is a mana ability (CR 605.1a) and skips the stack. Its rider schedules the
// delayed draw (CR 603.7a, next_upkeep_draw.go) for the controller as the mana is
// added; it happens even though the Sextant is gone by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97ba11a1-16cd-47ee-a187-1e919c468e4e",
		Name:         "Barbed Sextant",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}, Sacrifice this artifact: Add one mana of any color. Draw a card at the beginning of the next turn's upkeep.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				ctx := effects.NewContext(g, &game.StackItem{Controller: controller, SourceCardID: source})
				return effects.DrawACardAtTheNextUpkeep("Barbed Sextant — draw a card").Apply(ctx)
			},
		}},
	})
}
