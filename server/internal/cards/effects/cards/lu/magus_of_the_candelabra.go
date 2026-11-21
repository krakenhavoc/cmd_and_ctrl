package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Candelabra — Creature — Human Wizard {G}, 1/2:
//
//	"{X}, {T}: Untap X target lands."
//
// The target clause is exactly X lands (CountFromX, Crackle with Power's
// shape with "exactly" for "up to"): the activator announces X with the
// activation (CR 602.2b), chooses that many distinct lands as it goes on the
// stack, and each one still legal as the ability resolves is untapped
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30e58e88-ac99-4d32-b2fe-6d6cb6288af3",
		Name:         "Magus of the Candelabra",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Untap X target lands.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: magusOfTheCandelabraTargets(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

func magusOfTheCandelabraTargets() *game.TargetSpec {
	spec := effects.TargetPermanent("X target lands", effects.Land())
	spec.CountFromX = true
	return spec
}
