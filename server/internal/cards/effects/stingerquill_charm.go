package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stingerquill Charm — Instant {B}{R} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Stingerquill Charm deals 3 damage to any target.
//	 • Target creature gains first strike and deathtouch until end of turn.
//	 • Create a 2/2 colorless Wizard Soldier creature token named Cadet.
//	   It gains haste until end of turn."
//
// The haste is an until-end-of-turn grant on the new token, not a
// printed keyword of the template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "987aadac-d6a6-4e50-9468-9e8e53cc5529",
		Name:         "Stingerquill Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Stingerquill Charm deals 3 damage to any target.", TargetAny(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				}),
			ModeDoing("Target creature gains first strike and deathtouch until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "deathtouch"}, Label: "Stingerquill Charm"}.Apply(ctx)
				}),
			ModeDoing("Create a 2/2 colorless Wizard Soldier creature token named Cadet. It gains haste until end of turn.",
				nil, rfSpellBCadetWithHaste),
		),
	})
}
