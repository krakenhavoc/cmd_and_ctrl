package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jiang Yanggu, Wildcrafter — Legendary Planeswalker — Yanggu {2}{G}, loyalty 3
// (Foundations Jumpstart, #5929):
//
//	"Each creature you control with a +1/+1 counter on it has "{T}: Add one
//	 mana of any color."
//	 −1: Put a +1/+1 counter on target creature."
//
// Chromatic Lantern's layer-6 grant (ADR 0093) over the creatures its
// controller controls that have a +1/+1 counter on them, read live each time
// the layers are computed, so a creature gains the ability the moment a
// counter lands and loses it when the last one goes. The granted ability is a
// mana ability that taps its creature, so CR 302.6 still keeps a creature
// that came under its controller's control this turn from using it. The −1
// puts a counter on any creature, whoever controls it.
//
// No simplification.
const jiangYangguGrant = "jiang-yanggu-wildcrafter/any-color"

func init() {
	Register(Spec{
		OracleID:        "04f2c320-dfcc-440a-9e24-3c2083d74e7c",
		Name:            "Jiang Yanggu, Wildcrafter",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Grants:          []AbilityGrant{AnyColorManaGrant(jiangYangguGrant)},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(jiangYangguCountered, jiangYangguGrant),
				"Each creature you control with a +1/+1 counter on it has \"{T}: Add one mana of any color.\""),
		},
		Activated: []ActivatedAbility{{
			Label:   "−1: Put a +1/+1 counter on target creature.",
			Cost:    LoyaltyCost(-1),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := firstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
	})
}

// jiangYangguCountered is "each creature you control with a +1/+1 counter on
// it".
func jiangYangguCountered(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		target.Counters[game.CounterPlusOne] > 0
}
