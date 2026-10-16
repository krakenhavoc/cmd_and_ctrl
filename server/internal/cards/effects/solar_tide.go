package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solar Tide — Sorcery {4}{W}{W}:
//
//	"Choose one —
//	 • Destroy all creatures with power 2 or less.
//	 • Destroy all creatures with power 3 or greater.
//	 Entwine—Sacrifice two lands. (Choose both if you pay the entwine
//	 cost.)"
//
// Entwine is CR 702.42 (#3866), here a non-mana cost: the two lands are
// named on sacrifice_ids and leave together with the spell on the stack
// (EntwineSacrifice). Each bullet is a DestroyAllMatching sweep, power
// read as it resolves. Entwined, the small creatures are destroyed
// first and the big ones after (CR 702.42b), so the second sweep reads
// the board the first one left: a creature that regenerated stays and
// is judged on its power then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0f7a7348-3081-4fbb-9386-719a4ae7fa57",
		Name:          "Solar Tide",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{EntwineSacrifice(2, "two lands", Land())},
		Modes: Entwining(ChooseOne(
			ModeWithPurpose(ModeDoing("Destroy all creatures with power 2 or less.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Creature(), PowerLE(2))}.Apply(ctx)
				}), solarTideSweep),
			ModeWithPurpose(ModeDoing("Destroy all creatures with power 3 or greater.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Creature(), PowerGE(3))}.Apply(ctx)
				}), solarTideSweep),
		)),
	})
}

// solarTideSweep is each bullet's purpose (ADR 0126 §6): a creature
// wipe that spares the creatures on the other side of its power line.
var solarTideSweep = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}
