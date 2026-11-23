package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// D'Ghor Bloodletter — Creature — Klingon Assassin {2}{B}, 2/2:
//
//	"When this creature enters, choose one —
//	 • All other creatures get -1/-1 until end of turn.
//	 • Each player sacrifices a creature of their choice."
//
// A modal enters trigger (CR 603.3c, Charming Scoundrel's shape). The
// first bullet's "other" is object identity: the Bloodletter that entered
// is spared, and the set of creatures is fixed as the trigger resolves
// (CR 611.2c). The second bullet is Innocent Blood's edict, and it
// includes the Bloodletter's controller, who may sacrifice the
// Bloodletter itself.
//
// No simplification.
func init() {
	bloodletter := effects.WhenThisEnters("D'Ghor Bloodletter — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bloodletter.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeDoing("All other creatures get -1/-1 until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.TrekB11OtherThanSourceObject(item)),
					Power:     -1,
					Toughness: -1,
					Label:     "D'Ghor Bloodletter — -1/-1",
				}.Apply(ctx)
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}}),
		effects.ModeDoing("Each player sacrifices a creature of their choice.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.EachPlayerSacrifices{Match: effects.Creature(), Label: "a creature"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "ae5ee92a-5915-49ce-aa4c-611ab711b2ef",
		Name:         "D'Ghor Bloodletter",
		Completeness: effects.CompletenessFull,
		// The second bullet is an edict (ADR 0142's AnswerRemove on the row).
		Triggered: []game.TriggeredAbility{effects.TriggerEdict(bloodletter)},
	})
}
