package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherflux Reservoir — Artifact {4} (EDHREC rank 368):
//
//	"Whenever you cast a spell, you gain 1 life for each spell you've
//	 cast this turn.
//	 Pay 50 life: This artifact deals 50 damage to any target."
//
// The trigger is a cast trigger (YouCast(nil)). The count is read as
// it resolves (CR 608.2h), off the per-turn cast tally, which the cast
// path bumps before EventCast: the spell that triggered it is counted,
// so the first spell of a turn gains 1, the second 2, and a spell cast
// in response to the trigger is counted as well, as the printed "you've
// cast this turn" asks. A copy is not cast and neither triggers nor
// counts (CR 707.10).
//
// "Pay 50 life" is a cost (CR 119.4, CR 118.3): a player with less than
// 50 life can't activate it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f15125e-ae89-471e-b055-226a783bc7c1",
		Name:         "Aetherflux Reservoir",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(nil),
				"Aetherflux Reservoir — gain 1 life for each spell you've cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GainLife{Amount: g.CastTallyFor(ctx.Controller()).Total}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay 50 life: This artifact deals 50 damage to any target.",
			Cost:    PayLife(50),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 50)),
			Effect:  sourceDealsDamageToEachLegalTarget(50),
		}},
	})
}
