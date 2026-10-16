package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Field — Land (EDHREC top 1,000, #3654):
//
//	"{T}: Add {C}.
//	 {T}: Add {W} or {B}. Activate only if you control a Swamp."
//
// Two mana abilities on the one land, like Fanatic of Rhonas: the
// colourless one is always on, the coloured one carries the "Activate
// only if" gate as a Condition (CR 602.5), checked before the cost is
// paid. The Swamp is read by subtype, so a Swamp dual or a land that
// something else turned into a Swamp counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "439de49b-1091-4688-9ffb-80a025df31c2",
		Name:         "Tainted Field",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|B}",
				Label:    "Add {W} or {B}. Activate only if you control a Swamp.",
				Condition: func(g *game.Game, controller, source uuid.UUID) bool {
					return ControlsA(g.TextReaderOfForEffect(source).TextSubtype("Swamp"))(g, controller)
				},
			},
		},
	})
}
