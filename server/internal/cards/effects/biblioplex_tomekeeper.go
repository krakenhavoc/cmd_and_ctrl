package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Biblioplex Tomekeeper — Artifact Creature — Construct {4}, 3/4 (Secrets
// of Strixhaven):
//
//	"When this creature enters, choose up to one —
//	 • Target creature becomes prepared. (Only creatures with prepare
//	 spells can become prepared.)
//	 • Target creature becomes unprepared."
//
// A real "choose up to one" (Ertai Resurrected's shape): each mode names
// its own target, so choosing none puts nothing on the stack's targets.
// Both instructions are the engine's own (BecomePrepared, BecomeUnprepared,
// CR 722.3): a creature with no prepare spell stays as it is, and one that
// is not prepared stays not prepared.
func init() {
	Register(Spec{
		OracleID:     "08a0e845-9ef9-4de1-b9fe-4530357346ca",
		Name:         "Biblioplex Tomekeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Key:       socB20TomekeeperLabel,
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseN("Choose up to one", 0, 1,
				ModeDoing("Target creature becomes prepared.",
					TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return BecomePrepared{Target: t.ID}.Apply(ctx)
					}),
				ModeDoing("Target creature becomes unprepared.",
					TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return BecomeUnprepared{Target: t.ID}.Apply(ctx)
					}),
			),
		}},
	})
}
