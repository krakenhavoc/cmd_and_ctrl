package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waxing Moon — Instant {1}{G} (Eldritch Moon, S82 OLD, #5541):
//
//	"Transform up to one target Werewolf you control. Creatures you control
//	 gain trample until end of turn."
//
// The transform is the target's turning over in place (a target that has left
// or been illegal is dropped by the CR 608.2b re-check, and a creature that is
// not a double-faced card simply does not transform). The grant is Crash
// Through's: the creatures you control as the spell resolves, including one
// just turned over, gain trample until end of turn. The grant happens with no
// target chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbb4f08f-3016-40c2-ae7c-afc720eeb27a",
		Name:         "Waxing Moon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to one target Werewolf you control", effects.YouControl(), effects.OfSubtype("Werewolf")).WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if err := (effects.Transform{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"trample"},
				Label:    "Waxing Moon — creatures you control gain trample",
			},
		),
	})
}
