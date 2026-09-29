package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Earth Crystal — Legendary Artifact {2}{G}{G}:
//
//	"Green spells you cast cost {1} less to cast.
//	 If one or more +1/+1 counters would be put on a creature you
//	 control, twice that many +1/+1 counters are put on that creature
//	 instead.
//	 {4}{G}{G}, {T}: Distribute two +1/+1 counters among one or two
//	 target creatures you control."
//
// The green Crystal — see the_water_crystal.go / the_wind_crystal.go
// for the shape's other two members. The cost clause is the ordinary
// ColoredSpell discount. The doubler reuses
// plusOneCounterPlacementOnYourCreature, Hardened Scales' and
// Branching Evolution's shared "+1/+1 counter placed on a creature you
// control" predicate (helpers.go), with ×2 instead of +1 — this
// Crystal doubles rather than bumps. The activated ability is
// Armament Dragon's bounded-distribution shape (TargetCreature +
// WithCount + Dividing + PutDividedCounters), one target fewer.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5e7ef7fe-968b-4ada-9fe4-6fda0541aafc",
		Name:         "The Earth Crystal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Green spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("G")),
		},
		Replacements: []game.ReplacementEffect{
			{
				Watches:   []game.EventKind{game.EventCounterPlaced},
				AppliesTo: plusOneCounterPlacementOnYourCreature,
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.CounterDelta *= 2
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "The Earth Crystal: double +1/+1 counters",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{G}{G}, {T}: Distribute two +1/+1 counters among one or two target creatures you control.",
			Cost:  Plus(ManaCost("{4}{G}{G}"), TapCost()),
			Targets: TargetCreature("one or two target creatures you control", YouControl()).
				WithCount(1, 2).Dividing(Divide(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
