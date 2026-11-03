package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandman's Quicksand — Sorcery {1}{B}{B} (Marvel's Spider-Man, #4460):
//
//	"Mayhem {3}{B} (You may cast this card from your graveyard for {3}{B}
//	 if you discarded it this turn. Timing rules still apply.)
//	 All creatures get -2/-2 until end of turn. If this spell's mayhem
//	 cost was paid, creatures your opponents control get -2/-2 until end
//	 of turn instead."
//
// Golden Demise's shape with mayhem's alternative cost (CR 702.187) as
// the condition. The set of creatures is locked as the effect begins
// (CR 611.2c), as BoostUntilEOT does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f2b376d2-a2f4-4d23-a1fb-eb7b6ebd0b7a",
		Name:             "Sandman's Quicksand",
		Purpose:          game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness:     CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{3}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			match := Creature()
			label := "Sandman's Quicksand — all creatures get -2/-2"
			if ctx.Item != nil && PaidMayhem(ctx.Item.AltCost) {
				match = And(Creature(), OpponentControls())
				label = "Sandman's Quicksand — creatures your opponents control get -2/-2"
			}
			return BoostUntilEOT{Match: match, Power: -2, Toughness: -2, Label: label}.Apply(ctx)
		},
	})
}
