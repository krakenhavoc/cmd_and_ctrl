package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nexus Mentality — Instant {3}{U} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Choose one. If you control a commander as you cast this spell, you
//	 may choose both instead.
//	 • Move all counters from target nonland permanent you control onto
//	   another target nonland permanent you control.
//	 • Remove all counters from target nonland permanent you control.
//	   Draw a card for each counter removed this way."
//
// The mode count is the Commander Legends Will cycle's
// (OrUpToIf(2, YouControlACommander)): judged as the spell is cast,
// anybody's commander counts, and losing it afterwards changes nothing
// (the rulings). The modes resolve in printed order (CR 608.2c).
//
// Moving a counter is removing it from one permanent and putting it on
// the other (CR 122.5, the ruling), so both halves are ordinary counter
// events and a "put" doubler sees the placement. If either target is no
// longer legal the move is not possible and nothing moves (CR 122.5).
// The second mode draws one card for each counter that really came off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bdbe36f-b46d-4746-baf2-04f0977bb532",
		Name:         "Nexus Mentality",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Move all counters from target nonland permanent you control onto another target nonland permanent you control.",
				Clauses(
					PermanentYouControl("target nonland permanent you control", Nonland()),
					Distinct(PermanentYouControl("another target nonland permanent you control", Nonland())),
				),
				nexusMentalityMove),
			ModeDoing("Remove all counters from target nonland permanent you control. Draw a card for each counter removed this way.",
				PermanentYouControl("target nonland permanent you control", Nonland()),
				nexusMentalityRemoveAndDraw),
		).OrUpToIf(2, YouControlACommander),
	})
}

// nexusMentalityMove is the first bullet.
func nexusMentalityMove(_ *game.StackItem, ctx *Context, occ int) error {
	var from, to uuid.UUID
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		switch t.Slot {
		case 0:
			from = t.ID
		case 1:
			to = t.ID
		}
	}
	if from == uuid.Nil || to == uuid.Nil || from == to {
		return nil
	}
	src, ok := ctx.Game.LookupCardForEffect(from)
	if !ok {
		return nil
	}
	have := socB9CounterCounts(src)
	for _, kind := range socB9CounterKinds(src) {
		removed, err := socB9RemoveCounters(ctx, from, kind, have[kind])
		if err != nil {
			return err
		}
		if removed <= 0 {
			continue
		}
		if err := (AddCounter{Target: to, Kind: kind, N: removed}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// nexusMentalityRemoveAndDraw is the second bullet.
func nexusMentalityRemoveAndDraw(_ *game.StackItem, ctx *Context, occ int) error {
	var perm uuid.UUID
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			perm = t.ID
			break
		}
	}
	if perm == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(perm)
	if !ok {
		return nil
	}
	have := socB9CounterCounts(c)
	total := 0
	for _, kind := range socB9CounterKinds(c) {
		removed, err := socB9RemoveCounters(ctx, perm, kind, have[kind])
		if err != nil {
			return err
		}
		total += removed
	}
	if total <= 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: total}.Apply(ctx)
}
