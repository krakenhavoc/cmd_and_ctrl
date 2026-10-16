package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wastewood Verge — Land (EDHREC top 1,000, #3654):
//
//	"{T}: Add {G}.
//	 {T}: Add {B}. Activate only if you control a Swamp or a Forest."
//
// A Verge: the first colour is always on, the second is gated by an
// "Activate only if you control" Condition (CR 602.5) that reads
// subtypes, so a Forest or Swamp dual or a land that has been made one
// counts. Like Tainted Wood, one ManaAbility per printed ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6e0574c-3e2b-4c40-b17a-05bce3d49309",
		Name:         "Wastewood Verge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}. Activate only if you control a Swamp or a Forest.",
				Condition: func(g *game.Game, controller, source uuid.UUID) bool {
					text := g.TextReaderOfForEffect(source)
					return ControlsA(text.TextSubtype("Swamp"))(g, controller) || ControlsA(text.TextSubtype("Forest"))(g, controller)
				},
			},
		},
	})
}
