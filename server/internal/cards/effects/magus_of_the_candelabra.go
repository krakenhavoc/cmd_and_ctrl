package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "30e58e88-ac99-4d32-b2fe-6d6cb6288af3",
		Name:         "Magus of the Candelabra",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Untap X target lands.",
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Targets: magusOfTheCandelabraTargets(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (UntapTarget{Target: t.ID}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

func magusOfTheCandelabraTargets() *game.TargetSpec {
	spec := TargetPermanent("X target lands", Land())
	spec.CountFromX = true
	return spec
}
