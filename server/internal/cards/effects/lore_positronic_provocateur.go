package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lore, Positronic Provocateur — Legendary Artifact Creature — Android
// Rogue {1}{U/B}{U/B}, 2/3 (Star Trek, #4973):
//
//	"When Lore enters, choose one —
//	 • Target creature gets +1/+0 until end of turn and can't be blocked
//	   this turn.
//	 • Surveil 2."
//
// A modal ETB (CR 603.3c), Charming Prince's shape: the mode is chosen as
// the trigger is put on the stack, and only the first bullet has a target
// clause. That bullet is TVA Bureaucrat's pump and evasion pointed at the
// target, both skipped when it is no longer legal (CR 608.2b). The second
// is the ordinary surveil (CR 701.25).
//
// No simplification.
func init() {
	lore := WhenThisEnters("Lore, Positronic Provocateur — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	lore.Modes = ChooseOne(
		ModeDoing("Target creature gets +1/+0 until end of turn and can't be blocked this turn.",
			TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Label: "Lore, Positronic Provocateur — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked,
					Label: "Lore, Positronic Provocateur — can't be blocked"}.Apply(ctx)
			}),
		ModeDoing("Surveil 2.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Surveil{Player: item.Controller, N: 2}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "ee712451-01de-4486-9048-fafb5aa5a584",
		Name:         "Lore, Positronic Provocateur",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{lore},
	})
}
