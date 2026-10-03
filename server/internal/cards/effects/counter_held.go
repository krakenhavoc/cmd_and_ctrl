package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_held.go — "for as long as it has a <kind> counter on it"
// (ADR 0109 §2, #1604):
//
//	"Put a flood counter on target land. That land is an Island in
//	 addition to its other types for as long as it has a flood counter
//	 on it."                                            Aquitect's Will
//	"{3}{B}, {T}: Put a shadow counter on target creature. For as long
//	 as that creature has a shadow counter on it, it's a Wraith in
//	 addition to its other types."       Minas Morgul, Dark Fortress
//
// A continuous effect from a resolving spell or ability (CR 611.2) whose
// duration is the counter (CR 611.2b): game.WhilePinnedHasCounter,
// pinned to the permanent the counter went on. It ends the moment the
// last counter of that kind leaves, and for good: a new counter later
// does not bring it back. The source can be long gone; the effect is
// about the permanent with the counter.
//
// The counter goes on first and the duration is built second, in the
// order the card prints them. A counter that could not be put on (the
// permanent left, or a replacement stopped it) means the duration never
// starts, so nothing else happens (CR 611.2b).

// DurationWhileItHasCounter is "for as long as <object> has a <kind>
// counter on it". False when the duration would never start.
func DurationWhileItHasCounter(ctx *Context, object uuid.UUID, kind string) (game.Duration, bool) {
	return ctx.Game.ForAsLongAsPinnedHasCounterDuration(object, kind)
}

// CounterThenWhileItHasIt puts a <kind> counter on `object`, then gives
// it `mods` for as long as it has a counter of that kind: Aquitect's
// Will's Island, Liege of the Tangle's 8/8 Elemental, Minas Morgul's
// Wraith.
func CounterThenWhileItHasIt(ctx *Context, object uuid.UUID, kind, label string, mods ...game.Mod) error {
	if object == uuid.Nil {
		return nil
	}
	if err := (AddCounter{Target: object, Kind: kind, N: 1}).Apply(ctx); err != nil {
		return err
	}
	d, ok := DurationWhileItHasCounter(ctx, object, kind)
	if !ok {
		return nil
	}
	return ScopedEffectFor{Target: object, Mods: mods, Duration: d, Label: label}.Apply(ctx)
}

// FloodTargetLand is "put a flood counter on target land. That land is
// an Island in addition to its other types for as long as it has a
// flood counter on it" (Aquitect's Will, Xolatoyac). "In addition to its
// other types" takes nothing away (CR 205.1b, 305.7): the land keeps its
// own land types and abilities and also taps for {U}.
func FloodTargetLand(ctx *Context, name string) error {
	return CounterThenWhileItHasIt(ctx, FirstLegalBattlefieldTarget(ctx), "flood",
		name+" — an Island in addition to its other types while it has a flood counter",
		game.AddSubtypesMod("Island"))
}

// GrantWhileItHasCounter puts a <kind> counter on `object`, then gives
// it the ability bundles `keys` for as long as it has a counter of that
// kind: Makeshift Mannequin, Mathas, Obsidian Fireheart. The bundles are
// declared in the card's Spec.Grants (ADR 0093 PR 4).
func GrantWhileItHasCounter(ctx *Context, object uuid.UUID, kind, label string, keys ...string) error {
	if object == uuid.Nil {
		return nil
	}
	if err := (AddCounter{Target: object, Kind: kind, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return grantWhileItHasCounter(ctx, object, kind, label, keys...)
}

// grantWhileItHasCounter is GrantWhileItHasCounter for a permanent that
// already has the counter (Makeshift Mannequin's entered with it).
func grantWhileItHasCounter(ctx *Context, object uuid.UUID, kind, label string, keys ...string) error {
	d, ok := DurationWhileItHasCounter(ctx, object, kind)
	if !ok {
		return nil
	}
	return GrantAbilitiesFor{Target: object, Keys: keys, Duration: d, Label: label}.Apply(ctx)
}
