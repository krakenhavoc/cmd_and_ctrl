package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	titanium := effects.WheneverThisAttacks("Titanium Man — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	titanium.Modes = effects.ChooseOne(
		effects.ModeDoing("Titanium Man gains flying until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.SourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
					Label: "Titanium Man — flying"}.Apply(ctx)
			}),
		effects.ModeDoing("Titanium Man deals 1 damage to any target.", effects.TargetAny(),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "84544d35-7acf-4643-965d-66af90fae51b",
		Name:         "Titanium Man",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{titanium},
	})
}
