package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rain-Slicked Copse — Land — Forest Island (Secrets of Strixhaven
// Commander, #4834): "({T}: Add {G} or {U}.) This land enters tapped.
// Cycling {2}". The green-blue member of the cycling-land cycle, the
// same shape as Glittering Massif, Coastal Peak and Canyon Slough.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "633a41f0-3889-49a0-b2fc-9baf59f4ddc2",
		Name:          "Rain-Slicked Copse",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "U")},
	})
}
