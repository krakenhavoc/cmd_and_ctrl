package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "6bdbe36f-b46d-4746-baf2-04f0977bb532",
		Name:         "Nexus Mentality",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Move all counters from target nonland permanent you control onto another target nonland permanent you control.",
				effects.Clauses(
					effects.PermanentYouControl("target nonland permanent you control", effects.Nonland()),
					effects.Distinct(effects.PermanentYouControl("another target nonland permanent you control", effects.Nonland())),
				),
				nexusMentalityMove),
			effects.ModeDoing("Remove all counters from target nonland permanent you control. Draw a card for each counter removed this way.",
				effects.PermanentYouControl("target nonland permanent you control", effects.Nonland()),
				nexusMentalityRemoveAndDraw),
		).OrUpToIf(2, effects.YouControlACommander),
	})
}

// nexusMentalityMove is the first bullet.
func nexusMentalityMove(_ *game.StackItem, ctx *effects.Context, occ int) error {
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
	have := effects.SocB9CounterCounts(src)
	for _, kind := range effects.SocB9CounterKinds(src) {
		removed, err := effects.SocB9RemoveCounters(ctx, from, kind, have[kind])
		if err != nil {
			return err
		}
		if removed <= 0 {
			continue
		}
		if err := (effects.AddCounter{Target: to, Kind: kind, N: removed}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// nexusMentalityRemoveAndDraw is the second bullet.
func nexusMentalityRemoveAndDraw(_ *game.StackItem, ctx *effects.Context, occ int) error {
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
	have := effects.SocB9CounterCounts(c)
	total := 0
	for _, kind := range effects.SocB9CounterKinds(c) {
		removed, err := effects.SocB9RemoveCounters(ctx, perm, kind, have[kind])
		if err != nil {
			return err
		}
		total += removed
	}
	if total <= 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: total}.Apply(ctx)
}
