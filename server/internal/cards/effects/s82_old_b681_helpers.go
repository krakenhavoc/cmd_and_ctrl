package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b681_helpers.go — Mercadian Masques, "Activated abilities" (S82 OLD
// batch 681, #5619).

// oldB681DestroyTheObjectBody is a delayed trigger that destroys the object it
// names, if it is still that object on the battlefield (CR 603.7c). An
// on-disk key: never renamed, never reused.
var oldB681DestroyTheObjectBody = game.DelayedBody("destroy/old-b681-the-object", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !OnBattlefield(g, p.Object.ID) {
		return nil
	}
	return DestroyTarget{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// OldB681PumpThenDestroyAtEndStep is Puffer Extract's resolution: the target
// gets +X/+X until end of turn and is destroyed at the beginning of the next
// end step.
func OldB681PumpThenDestroyAtEndStep(ctx *Context, name string) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: ctx.X(), Toughness: ctx.X(), Label: name + " — +X/+X"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  name + " — destroy " + c.Name,
		Cards:  []uuid.UUID{id},
		Body:   oldB681DestroyTheObjectBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}

// OldB681AfterBlockersAreDeclared is "Activate only during combat after
// blockers are declared" (Trap Runner): the declare blockers step and the
// combat steps after it.
func OldB681AfterBlockersAreDeclared(g *game.Game, _, _ uuid.UUID) bool {
	switch g.Turn.Step {
	case game.StepDeclareBlockers, game.StepFirstStrikeDamage, game.StepCombatDamage, game.StepEndCombat:
		return true
	}
	return false
}
