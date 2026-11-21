package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulna Alley Shopkeep — Creature — Goblin Warlock {2}{B}, 2/3 (Secrets of
// Strixhaven, #4456):
//
//	"Menace
//	 Infusion — This creature gets +2/+0 as long as you gained life this
//	 turn."
//
// Menace is a printed keyword. Infusion is an ability word (CR 207.2c) over a
// live layer-7c static on this creature alone: it applies the moment you
// gain life (the layer cache is told about life changes, DependsOnLifeTotal)
// and stops when the turn ends.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9da7f9ec-872a-471d-8d60-5f4213f0204f",
		Name:            "Ulna Alley Shopkeep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			effects.SocB38InfusionStatic(effects.B16Anthem(effects.SocB38InfusionAppliesToThis, 2, 0), "Infusion — this creature gets +2/+0 as long as you gained life this turn."),
		},
	})
}
