package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titanium Man — {2}{B}{B} Legendary Artifact Creature — Human Villain 3/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Titanium Man attacks, choose one —
//	 • Titanium Man gains flying until end of turn.
//	 • Titanium Man deals 1 damage to any target."
//
// The mode is chosen as the trigger is put on the stack; only the damage
// mode has a target (CR 700.2b).
//
// No simplification.
func init() {
	titanium := WheneverThisAttacks("Titanium Man — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	titanium.Modes = ChooseOne(
		ModeDoing("Titanium Man gains flying until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if !sourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
					Label: "Titanium Man — flying"}.Apply(ctx)
			}),
		ModeDoing("Titanium Man deals 1 damage to any target.", TargetAny(),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "84544d35-7acf-4643-965d-66af90fae51b",
		Name:         "Titanium Man",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{titanium},
	})
}
