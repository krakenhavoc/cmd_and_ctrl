package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspired Inventor — Creature — Human Artificer {2}{W}, 2/2:
//
//	"When this creature enters, choose one —
//	 • You get {E}{E}{E} (three energy counters).
//	 • Put a +1/+1 counter on target creature.
//	 • Create a 1/1 colorless Servo artifact creature token."
//
// ADR 0129 PR 1. A modal trigger: the mode is chosen as it goes on the
// stack (CR 603.3c), then the target for the counter mode (CR 603.3d).
//
// No simplification.
func init() {
	inventor := WhenThisEnters("Inspired Inventor — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	inventor.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("You get {E}{E}{E}.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return GetEnergy{N: 3}.Apply(ctx)
		}), game.Purpose{Energy: 3}),
		ModeDoing("Put a +1/+1 counter on target creature.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("Create a 1/1 colorless Servo artifact creature token.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1}),
	)
	Register(Spec{
		OracleID:     "126679aa-f150-4949-8566-12d223af3960",
		Name:         "Inspired Inventor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{inventor},
	})
}
