package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whisperer of the Wilds — Creature — Human Shaman {1}{G}, 0/2:
//
//	"{T}: Add {G}.
//	 Ferocious — {T}: Add {G}{G}. Activate only if you control a creature
//	 with power 4 or greater."
//
// Fanatic of Rhonas's two mana abilities, the second gated by the
// ferocious condition before any cost is paid (CR 602.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1e456b6-5d2b-4c42-9464-54a8f9f8f3ba",
		Name:         "Whisperer of the Wilds",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}{G}",
				Label:    "Ferocious — Add {G}{G}",
				Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
					return effects.B24ControlsCreatureWithPowerAtLeast(g, controller, 4)
				},
			},
		},
	})
}
