package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Willowrush Verge — Land (Aetherdrift, #4464):
//
//	"{T}: Add {U}.
//	 {T}: Add {G}. Activate only if you control a Forest or an Island."
//
// Wastewood Verge's shape: the first colour is always on, the second is
// gated by an "Activate only if you control" condition that reads
// subtypes through the land's own text, so a Forest or Island dual, or a
// land made one, counts and the Verge itself never does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c52eaa87-9251-4a47-83fd-04e582ade612",
		Name:         "Willowrush Verge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U}",
				Label:    "Add {U}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}. Activate only if you control a Forest or an Island.",
				Condition: func(g *game.Game, controller, source uuid.UUID) bool {
					text := g.TextReaderOfForEffect(source)
					return ControlsA(text.TextSubtype("Forest"))(g, controller) || ControlsA(text.TextSubtype("Island"))(g, controller)
				},
			},
		},
	})
}
