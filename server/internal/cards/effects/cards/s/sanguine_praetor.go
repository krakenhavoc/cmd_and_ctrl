package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Praetor — Creature — Avatar Praetor {6}{B}{B}, 7/5 (Guildpact,
// S82 OLD batch 573, #5599):
//
//	"{B}, Sacrifice a creature: Destroy each creature with the same mana
//	 value as the sacrificed creature."
//
// The sacrificed creature's mana value is read as it last existed on the
// battlefield (CR 608.2h; a token that is no copy is 0), and the destruction
// is one simultaneous sweep (CR 701.8) that also takes the Praetor itself
// when its mana value matches.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af6124cf-e8c5-44f3-964a-4ac0caa6f22b",
		Name:         "Sanguine Praetor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Destroy each creature with the same mana value as the sacrificed creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeN(1, "a creature", effects.Creature())),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyAllMatching{
					Match: effects.And(effects.Creature(), effects.ManaValueIs(effects.SacrificedManaValue(ctx))),
				}.Apply(ctx)
			})),
		}},
	})
}
