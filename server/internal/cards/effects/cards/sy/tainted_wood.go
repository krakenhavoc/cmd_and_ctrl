package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Wood — Land (EDHREC top 1,000, #3654):
//
//	"{T}: Add {C}.
//	 {T}: Add {B} or {G}. Activate only if you control a Swamp."
//
// Two mana abilities on the one land, like Fanatic of Rhonas: the
// colourless one is always on, the coloured one carries the "Activate
// only if" gate as a Condition (CR 602.5), checked before the cost is
// paid. The Swamp is read by subtype, so a Swamp dual or a land that
// something else turned into a Swamp counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa6d05a1-3df4-4751-b1a0-8d9693faec73",
		Name:         "Tainted Wood",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{B|G}",
				Label:    "Add {B} or {G}. Activate only if you control a Swamp.",
				Condition: func(g *game.Game, controller, source uuid.UUID) bool {
					return effects.ControlsA(g.TextReaderOfForEffect(source).TextSubtype("Swamp"))(g, controller)
				},
			},
		},
	})
}
