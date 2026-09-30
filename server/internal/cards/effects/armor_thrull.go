package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armor Thrull — Creature — Thrull {2}{B}, 1/3 (#1664):
//
//	"{T}, Sacrifice this creature: Put a +1/+2 counter on target
//	 creature."
//
// Waited on #1664: a +1/+2 counter used to be stored and change
// nothing. Every P/T counter kind now counts (game.PTCounterDelta,
// CR 122.1a). The sacrifice is a cost, so it is paid at announce and
// the Thrull's dies-triggers resolve above the ability. Hardened
// Scales does not add to it (that card says "+1/+1 counters"), and a
// Doubling Season makes it two +1/+2 counters on a creature you
// control. The {T} is a creature's {T}: summoning sickness applies
// (CR 302.6).
func init() {
	Register(Spec{
		OracleID:     "35453c9e-e1ae-4fe9-926d-75724deb0555",
		Name:         "Armor Thrull",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Put a +1/+2 counter on target creature.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: target, Kind: "+1/+2", N: 1}.Apply(ctx)
			},
		}},
	})
}
