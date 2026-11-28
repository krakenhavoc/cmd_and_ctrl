package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jabari's Banner — Artifact {2} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"{1}, {T}: Target creature gains flanking until end of turn. (Whenever a creature without
//	 flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)"
//
// Bearer of Memory's timed keyword grant, over any creature; flanking is cumulative (CR 702.25f),
// so a creature that already has it gets a second instance, which the engine counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79560d80-1745-4773-a5c4-4c8e98475135",
		Name:         "Jabari's Banner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains flanking until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flanking")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flanking"}, Label: "Jabari's Banner — flanking"}.Apply(ctx)
			},
		}},
	})
}
