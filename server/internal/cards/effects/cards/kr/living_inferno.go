package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Inferno — Creature — Elemental {6}{R}{R}, 8/5 (Guildpact):
//
//	"{T}: This creature deals damage equal to its power divided as you
//	 choose among any number of target creatures. Each of those creatures
//	 deals damage equal to its power to this creature."
//
// The amount is the Inferno's power as the ability is announced and the split
// is announced with the targets (CR 601.2d, 602.2b; Orca, Siege Demon's
// DivideSourcePower), each target at least 1. On resolution the divided
// damage is dealt from the Inferno, and then each creature that was dealt
// damage this way deals damage equal to its power to the Inferno. A target
// that is illegal on resolution takes nothing and deals nothing, and its
// share is not reassigned. State-based actions are not checked between the
// two halves, so a creature the Inferno just killed still hits back, with the
// power it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6783c46d-e4e5-4b43-9d97-6eae8aab4cb9",
		Name:         "Living Inferno",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to its power divided as you choose among any number of target creatures. Each of those creatures deals damage equal to its power to this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideBy(effects.DivideSourcePower)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.DealDividedDamage(ctx); err != nil {
					return err
				}
				return g.DamageInstanceForEffect(func() error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard || item.Distribution[t.ID] <= 0 {
							continue
						}
						power := effects.B43PowerNowOrLastKnown(g, t.ID)
						if power <= 0 {
							continue
						}
						if err := (effects.DealDamage{Source: t.ID, Target: ctx.Source(), Amount: power}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}
