package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bleeding Edge — Sorcery {1}{B}{B} (War of the Spark, #5520):
//
//	"Up to one target creature gets -2/-2 until end of turn. Amass Zombies 2."
//
// The -2/-2 goes to the target if there is one and it is still legal as the
// spell resolves (CR 608.2b); the amass happens regardless (CR 701.47a), a
// spell with no target still amassing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "308e551a-6d60-4a28-9e22-1181995ca068",
		Name:         "Bleeding Edge",
		Completeness: effects.CompletenessFull,
		Purpose:      bleedingEdgePurpose(),
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if target, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.BoostUntilEOT{Target: target, Power: -2, Toughness: -2, Label: "Bleeding Edge — -2/-2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Amass{Subtype: "Zombie", N: 2}.Apply(ctx)
		},
	})
}

// bleedingEdgePurpose is the -2/-2 on the target until end of turn, and the
// amass.
func bleedingEdgePurpose() game.Purpose {
	p := effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))
	p.Amass = 2
	return p
}
