package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// provoke.go is the "target creature blocks IT this turn if able"
// family (#1684, ADR 0045 amendment of 2026-09-28, Decision 59):
// Provoke (CR 702.39), Grappling Hook, Turntimber Basilisk. Each is a
// resolving effect that registers a BlocksAttackerUntilEOT record on
// the chosen creature naming ONE attacking object; the engine's block
// requirement search (game/block_requirements.go) does the rest.

// Provoke (CR 702.39a) — "Whenever this creature attacks, you may have
// target creature defending player controls untap and block it if
// able."
//
// Not in the closed PrintedKeywords / HasKeyword table
// (game/keywords.go): that table is for keywords the engine READS as a
// bare string (a flyer's evasion, a lifelinker's gain, a prowess or
// exalted trigger it derives), which another permanent can grant.
// Provoke is a triggered ability like any other, wired up on each
// source by this constructor, and nothing in the catalog grants it. Two instances trigger
// separately (CR 702.39b), which two calls of this constructor are.
//
// "It" is the object that attacked — the trigger's own event object,
// read off item.Trigger (ADR 0041 P9) rather than captured, so the
// ability is a restore point while it waits on the stack.
func Provoke() game.TriggeredAbility {
	return Optional(game.TriggeredAbility{
		Watches:     []game.EventKind{game.EventAttack},
		AppliesTo:   ThisAttacked,
		TargetsFrom: TargetCreatureDefendingPlayerControls,
		Key:         "Provoke — untap target creature defending player controls; it blocks this creature if able",
		Effect:      provokeUntapAndBlock,
	}, "Provoke — have target creature defending player controls untap and block this creature if able?")
}

// provokeUntapAndBlock is Provoke's effect: untap the target, then it
// blocks the provoker this turn if able.
func provokeUntapAndBlock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := g.UntapTargetForEffect(id); err != nil {
		return err
	}
	return BlocksAttackerUntilEOT{
		Blocker:  id,
		Attacker: attackingObjectOf(ctx),
		Label:    "Provoke — blocks the provoking creature if able",
	}.Apply(ctx)
}

// TargetBlocksTheAttacker is Grappling Hook's "you may have target
// creature block it this turn if able", on a "whenever equipped
// creature attacks" trigger: "it" is the attacking creature the
// trigger's event names (the equipped creature), not the Equipment.
func TargetBlocksTheAttacker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BlocksAttackerUntilEOT{
		Blocker:  id,
		Attacker: attackingObjectOf(ctx),
		Label:    "blocks the attacking creature if able",
	}.Apply(ctx)
}

// TargetBlocksThisCreature is Turntimber Basilisk's "you may have
// target creature block this creature this turn if able": "this
// creature" is the ability's source object, attacking or not yet — the
// requirement waits for it to attack this turn, and asks nothing if it
// never does.
func TargetBlocksThisCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return BlocksAttackerUntilEOT{
		Blocker:  id,
		Attacker: ref,
		Label:    "blocks this creature if able",
	}.Apply(ctx)
}

// attackingObjectOf is the attacking object an attack trigger is
// about: the event's object (CR 603.10 last-known information, so its
// epoch is the one it had when it attacked), falling back to the
// ability's source for an item that carries no trigger context.
func attackingObjectOf(ctx *Context) game.ObjectRef {
	if it := ctx.Item; it != nil && it.Trigger != nil && it.Trigger.Object != nil {
		return it.Trigger.Object.Ref()
	}
	ref, _ := ctx.SourceRef()
	return ref
}
