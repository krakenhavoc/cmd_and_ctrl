package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ten Wizards Mountain — Plane — Shenmeng (Commander Masters, #4870):
//
//	"Whenever you roll the planar die, put a +1/+1 counter on up to one
//	 target creature.
//	 Whenever chaos ensues, creatures you control gain flying until end
//	 of turn."
//
// A plane (ADR 0216 §6). The first ability triggers on the planar controller
// rolling the planar die, whatever face it shows (CR 901.9: a blank, chaos
// or planeswalk), and an ordinary die roll is not one (the planar face is
// what marks the roll). When the face is the planeswalk symbol the
// ability still resolves after the plane is turned face down, as any
// independent ability does. The chaos ability (CR 311.7) gives the planar
// controller's creatures flying; the set is fixed as it resolves (CR
// 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0cb118e-6af1-41d2-847f-d685ced2165d",
		Name:         "Ten Wizards Mountain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventRollDie, youRolledThePlanarDie, "Ten Wizards Mountain — put a +1/+1 counter on up to one target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: t, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}), TargetCreature("up to one target creature").WithCount(0, 1)),
			WheneverChaosEnsues("Ten Wizards Mountain — creatures you control gain flying until end of turn",
				Do(GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"flying"}, Label: "Ten Wizards Mountain — flying"})),
		},
	})
}

// youRolledThePlanarDie is "whenever you roll the planar die" (CR 901.9):
// the planar controller finished a planar die roll, any face.
func youRolledThePlanarDie(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.PlanarFace != "" && ev.Actor == source.Controller
}
