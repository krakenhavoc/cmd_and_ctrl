package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcish Squatters — Creature — Orc {4}{R}, 2/3:
//
//	"Whenever this creature attacks and isn't blocked, you may gain control of target land
//	 defending player controls for as long as you control this creature. If you do, this
//	 creature assigns no combat damage this turn."
//
// #3531: the duration is "for as long as you control this creature" (CR 611.2b); a Squatters that
// left and returned before the trigger resolved is not the object the ability names, so nothing
// is taken.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ce936736-915f-44bb-a709-b3bdcf90a3be",
		Name:         "Orcish Squatters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(unblockedAttackerMay("Orcish Squatters", "gain control of target land defending player controls", "orcish-squatters-gain-control",
				func(ctx *Context, target uuid.UUID) error {
					d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return GainControl{Target: target, Duration: d, Label: "Orcish Squatters — gain control of the land"}.Apply(ctx)
				}), TargetOfDefendingPlayer("land", "target land defending player controls", Land())),
		},
	})
}
