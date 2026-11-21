package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Vatkeeper — Creature — Phyrexian Elf {1}{B}{G}, 3/3 (March of
// the Machine):
//
//	"When this creature enters, incubate 2.
//	 {5}: Transform target Incubator token you control. Double the
//	 number of +1/+1 counters on it."
//
// The ability's target is an Incubator token you control, front face up
// (IsIncubator): a token that already turned over is a Phyrexian Token,
// not an Incubator token, so it is not a legal target. The doubling is
// the second sentence of the SAME ability and so happens whether or not
// the transform did anything. It puts as many +1/+1 counters on the
// token as it has, through the counter pipeline, so Hardened Scales
// and Doubling Season apply to the doubling as printed. A token that
// has left, or was turned over in response, makes the target illegal
// and the whole ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b881d61c-d7cf-49af-bed7-3bc657ea1208",
		Name:         "Elvish Vatkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elvish Vatkeeper — incubate 2", effects.Do(effects.Incubate{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: Transform target Incubator token you control. Double the number of +1/+1 counters on it.",
			Cost:    effects.ManaCost("{5}"),
			Targets: effects.TargetIncubatorYouControl(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.Transform{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.B23DoublePlusOneCountersOn(ctx, t.ID)
				}
				return nil
			},
		}},
	})
}
