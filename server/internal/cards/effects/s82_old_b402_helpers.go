package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b402_helpers.go — Dragon's Maze fuse cards, S82 OLD batch 402 (#5564).

// S82OldB402OneSidedBiteAtAnyTarget is "target creature you control deals
// damage equal to its power to any target" as a spell's whole resolution
// (Blood): OneSidedBite's two-clause read with the second slot free to be a
// player or a planeswalker as well as a creature. Slot 0 deals, its power read
// as the spell resolves; slot 1 takes. Nothing is dealt unless both picks are
// still legal (CR 608.2b).
func S82OldB402OneSidedBiteAtAnyTarget(_ *game.StackItem, ctx *Context) error {
	biter, ok := ctx.ClauseTarget(0)
	if !ok || biter.Kind != game.TargetCard {
		return nil
	}
	victim, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(biter.ID)
	if !ok {
		return nil
	}
	return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
}

// s82OldB402EachCreatureEntered is the permanent an EventETB names, while it is
// a creature under any player's control (Beck's "whenever a creature enters
// this turn"). Types are read as the permanent is on the battlefield (CR 603.10).
func s82OldB402EachCreatureEntered(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// Beck's trigger condition and body. On-disk keys: never renamed, never reused.
var (
	S82OldB402AnyCreatureEntersCondition = game.DelayedEachCondition("etb/each-creature-any-player", s82OldB402EachCreatureEntered)
	S82OldB402DrawACardBody              = game.SimpleDelayedBody("beck-call/draw-a-card", Do(DrawCards{N: 1}))
)

// S82OldB402ReturnTargetCardToHand is "return target card from your graveyard to
// your hand" as a spell's whole resolution (Dirty): the first target, if it is
// still a legal card in the graveyard (CR 608.2b).
func S82OldB402ReturnTargetCardToHand(_ *game.StackItem, ctx *Context) error {
	pick, ok := ctx.ClauseTarget(0)
	if !ok || pick.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: pick.ID, Dest: game.ZoneHand}.Apply(ctx)
}

// S82OldB402CountersOnTheTarget is "put N <kind> counters on target creature" as
// a spell's whole resolution (Give's three +1/+1 counters): the first target card
// that is still legal as the spell resolves (CR 608.2b).
func S82OldB402CountersOnTheTarget(kind string, n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return AddCounter{Target: t.ID, Kind: kind, N: n}.Apply(ctx)
			}
		}
		return nil
	}
}
