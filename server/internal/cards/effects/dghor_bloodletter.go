package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	bloodletter := WhenThisEnters("D'Ghor Bloodletter — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bloodletter.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("All other creatures get -1/-1 until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return BoostUntilEOT{
					Match:     And(Creature(), trekB11OtherThanSourceObject(item)),
					Power:     -1,
					Toughness: -1,
					Label:     "D'Ghor Bloodletter — -1/-1",
				}.Apply(ctx)
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true}}),
		ModeDoing("Each player sacrifices a creature of their choice.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return EachPlayerSacrifices{Match: Creature(), Label: "a creature"}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "ae5ee92a-5915-49ce-aa4c-611ab711b2ef",
		Name:         "D'Ghor Bloodletter",
		Completeness: CompletenessFull,
		// The second bullet is an edict (ADR 0142's AnswerRemove on the row).
		Triggered: []game.TriggeredAbility{TriggerEdict(bloodletter)},
	})
}
