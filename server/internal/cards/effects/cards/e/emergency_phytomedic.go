package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergency Phytomedic // Seed Suture — Creature — Dryad Cleric {G/W},
// 1/1 // Sorcery {G/W} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Seed Suture — "Put a +1/+1 counter on target creature. You gain 1
//	 life."
//
// Seed Suture is shared with Blossom-Blessed Angel (seedSutureResolve).
//
// No simplification.
func init() {
	const id = "7737ccdc-49f2-463e-8ffd-43ab98d11f45"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Emergency Phytomedic",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Seed Suture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SeedSutureResolve,
	})
}
