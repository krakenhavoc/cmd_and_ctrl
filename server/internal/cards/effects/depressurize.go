package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Depressurize — Instant {1}{B} (Edge of Eternities, #4461):
//
//	"Target creature gets -3/-0 until end of turn. Then if that creature's
//	 power is 0 or less, destroy it."
//
// Two instructions in order (CR 608.2c). The -3/-0 is a layer 7c pump
// until end of turn; the power is then read through the layers with its
// counters and without clamping, so a creature that was already at 3 power
// or less is destroyed and a bigger one only shrinks. A creature that left
// in response is an illegal target and the spell does nothing (CR 608.2b).
// An indestructible creature keeps the -3/-0 and survives the destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abcb5321-c295-4abf-8622-7de79679d27a",
		Name:         "Depressurize",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove,
			Targets: game.ForTargets(HarmUntilEOT(0, -3, 0))},
		Targets:   TargetCreature("target creature"),
		OnResolve: depressurizeResolve,
	})
}

// depressurizeResolve shrinks the target, then destroys it if its power is
// now 0 or less.
func depressurizeResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: -3, Label: "Depressurize — -3/-0 until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !onBattlefield(ctx.Game, id) || c.PowerForComparison() > 0 {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}
