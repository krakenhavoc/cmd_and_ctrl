package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crenellated Wall — Artifact Creature — Wall {4}, 0/4 (Mercadian Masques, S82
// OLD, #5619):
//
//	"Defender
//	 {T}: Target creature gets +0/+4 until end of turn."
//
// Defender is printed data. The {T} is the Wall's own (summoning sickness
// applies, CR 302.6); the target is any creature, the Wall included, checked
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca9e41e8-7830-4721-b131-07d0b60297ad",
		Name:            "Crenellated Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +0/+4 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 4)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Toughness: 4, Label: "Crenellated Wall — +0/+4 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
