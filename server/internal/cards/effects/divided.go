package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// divided.go — #1563, CR 601.2d / 700.2i: "N damage divided as you
// choose among …". The clause declares the amount with Dividing; the
// caster announces the split with the targets; the engine refuses a
// split that gives a target 0 or does not add up; and resolution reads
// it back here. See game/divide.go and ADR 0065's 2026-09-27 amendment
// on divided effects.
//
//	Targets: TargetPermanent("any number of target creatures and/or planeswalkers",
//		Or(Creature(), Planeswalker())).WithCount(0, 4).Dividing(Divide(4)),       // Fury
//	Targets: TargetAny().WithCount(0, 0).Dividing(DivideX()),                        // Rolling Thunder
//	Targets: TargetPermanent(…).WithCount(0, 2).Dividing(DivideXDoublingFrom(6)),    // Shatterskull Smashing
//
// "Any number of targets" is WithCount(0, 0) or (0, N): the gate
// already refuses more targets than the amount, since each must be
// assigned at least 1, so the count needs no separate cap.

// Divide is a fixed amount divided as the caster chooses.
func Divide(total int) game.DivideSpec { return game.DivideSpec{Total: total} }

// DivideX is "X damage divided as you choose": the amount is the
// announced X. Only a spell's or an activated ability's clause may
// declare it — a trigger announces no X, and Register refuses it there.
func DivideX() game.DivideSpec { return game.DivideSpec{FromX: true} }

// DivideXDoublingFrom is DivideX that becomes twice X once X reaches
// `from` — Shatterskull Smashing's "If X is 6 or more, … deals twice X
// damage divided as you choose among them instead."
func DivideXDoublingFrom(from int) game.DivideSpec {
	return game.DivideSpec{FromX: true, DoubleFromX: from}
}

// DivideBy is an amount read by a registered rule at announce (#1657)
// rather than printed — "X is the number of lands you control",
// "damage equal to its power". The rules live in divide_amounts.go.
func DivideBy(rule game.DivideAmount) game.DivideSpec {
	return game.DivideSpec{AmountKey: rule}
}

// UpTo turns a division into "distribute UP TO that many" (Lathiel):
// the shares may sum to less than the amount, each chosen target still
// at least 1. UpTo(DivideBy(LifeYouGainedThisTurn)).
func UpTo(d game.DivideSpec) game.DivideSpec {
	d.UpTo = true
	return d
}

// DealDividedDamage deals each target of the resolving item its
// announced share of the division, from the item's source.
//
// Each target is re-checked first (CR 608.2b): one that left or stopped
// qualifying takes nothing, and its share is NOT redistributed — the
// others take exactly what was announced for them. The item's
// Distribution was settled at announce (a cast, an activation or a
// trigger's pick_target walk — the only three places a divided item's
// targets are chosen), so every legal target has a share of at least 1.
func DealDividedDamage(ctx *Context) error {
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	shares := ctx.Item.Distribution
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, t := range legal {
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: shares[t.ID]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// PutDividedCounters is the counters twin of DealDividedDamage —
// "Distribute two +1/+1 counters among one or two target creatures"
// (Abzan Charm). CR 601.2d governs "divide or distribute" alike: each
// target at least one, the announced split honoured, a departed
// target's counters lost rather than moved.
func PutDividedCounters(ctx *Context, kind string) error {
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	shares := ctx.Item.Distribution
	for _, t := range legal {
		if t.Kind != game.TargetCard || shares[t.ID] <= 0 {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: kind, N: shares[t.ID]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
