package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rain-Slicked Copse — Land — Forest Island (Secrets of Strixhaven
// Commander, #4834): "({T}: Add {G} or {U}.) This land enters tapped.
// Cycling {2}". The green-blue member of the cycling-land cycle, the
// same shape as Glittering Massif, Coastal Peak and Canyon Slough.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "633a41f0-3889-49a0-b2fc-9baf59f4ddc2",
		Name:          "Rain-Slicked Copse",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "U")},
	})
}
