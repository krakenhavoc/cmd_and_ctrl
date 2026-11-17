package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bane of the Living — Creature — Insect {2}{B}{B}, 4/3 (Legions, #5609):
//
//	"Morph {X}{B}{B}
//	 When this creature is turned face up, all creatures get -X/-X until
//	 end of turn."
//
// The X is chosen as the morph special action is taken (CR 107.3d), and
// CR 702.37f hands it to the turned-face-up trigger: the special action
// puts it on the turned-face-up event's Amount, which the trigger reads
// as it resolves (Aurelia's Vindicator's disguise X). The shrink is
// Toxic Deluge's: a layer-7c -X/-X over every creature on the
// battlefield as it resolves, the Bane itself included (CR 611.2c).
// Turned face up by an effect instead, nothing chose an X and the
// trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c99e4768-192c-49bf-8f51-c204a2248326",
		Name:             "Bane of the Living",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{X}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WhenThisIsTurnedFaceUp("Bane of the Living — all creatures get -X/-X until end of turn", baneOfTheLivingShrink),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}}),
		},
	})
}

// baneOfTheLivingShrink is "all creatures get -X/-X until end of turn",
// X the morph cost's X off the trigger's event.
func baneOfTheLivingShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := ctx.Trigger().Amount()
	if x <= 0 {
		return nil
	}
	return BoostUntilEOT{Match: Creature(), Power: -x, Toughness: -x, Label: "Bane of the Living — -X/-X"}.Apply(ctx)
}
