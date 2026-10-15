package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caldaia Strongarm — Creature — Human Warrior {4}{G}, 2/3:
//
//	"When this creature enters, put two +1/+1 counters on target
//	 creature.
//	 Blitz {3}{G}"
//
// Blitz (CR 702.152, #3772). The enters trigger targets any creature,
// the Strongarm itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e0e6938b-7abf-4f69-a3d6-82a58912bff4",
		Name:             "Caldaia Strongarm",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Caldaia Strongarm — two +1/+1 counters on target creature",
				twoPlusOneCountersOnFirstLegalTarget), TargetCreature("target creature")),
		},
	})
}

// twoPlusOneCountersOnFirstLegalTarget puts two +1/+1 counters on the
// trigger's target if it is still legal (CR 608.2b).
func twoPlusOneCountersOnFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return nothingIfGone(AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx))
		}
	}
	return nil
}
