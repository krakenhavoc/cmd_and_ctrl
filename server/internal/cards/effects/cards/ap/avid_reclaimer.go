package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avid Reclaimer — Creature — Human Druid {2}{G}, 2/2 (Hour of Devastation,
// S82 OLD batch 241, #5534):
//
//	"{T}: Add {G} or {U}. If you control a Nissa planeswalker, you gain 2
//	 life."
//
// One mana ability (CR 605.1a: it adds mana and has no target) whose rider
// gains 2 life when the mana is added if you control a planeswalker with the
// subtype Nissa.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a94bb381-74e2-488f-896d-a017a2b568ab",
		Name:         "Avid Reclaimer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}. If you control a Nissa planeswalker, you gain 2 life",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				if !effects.OldB241ControlsAPlaneswalkerOfSubtype(g, controller, "Nissa") {
					return nil
				}
				return effects.GainLifeRider(2)(g, controller, source)
			},
		}},
	})
}
