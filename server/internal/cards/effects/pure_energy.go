package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pure Energy — Sorcery {1}{W} (Star Trek, #4973):
//
//	"Distribute two +1/+1 counters among one or two target creatures.
//	 Beam me up {2}{W} (You may cast this card from your graveyard for
//	 {2}{W} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// Abzan Charm's divided counters (the split is announced with the
// targets, CR 601.2d) and beam me up from the graveyard
// (trekB4BeamMeUp).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f4831eee-fef6-4fef-a174-400a25d0837d",
		Name:             "Pure Energy",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB4BeamMeUp("{2}{W}")},
		Targets:          TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
