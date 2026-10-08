package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// awaken.go — the catalog side of awaken (CR 702.113, ADR 0135 §3,
// #2411). The verb is one engine function (`game.AwakenForEffect`);
// this file is the offer and the wrapper a card needs to print it.
//
//	Awaken N—[cost] (If you cast this spell for [cost], also put N
//	+1/+1 counters on target land you control and it becomes a 0/0
//	Elemental creature with haste. It's still a land.)
//
// A card declares it in three places, and the guard in Register keeps
// them together:
//
//	t := TargetCreatureOrPlaneswalker("target creature or planeswalker")
//	Register(Spec{
//		Targets:          t,
//		AlternativeCosts: []game.AlternativeCost{Awaken(4, "{5}{B}{B}", t)},
//		OnResolve:        AwakenAfter(4, destroyTheTarget),
//	})

// awakenKey is the offer's wire key, read back at resolution with
// ctx.PaidAltCost.
const awakenKey = "awaken"

// AwakenLandTargets is the clause awaken adds: "target land you
// control" (CR 702.113a). Not Distinct, so Earthen Arms may put its own
// two counters and its awaken counters on the same land (CR 601.2c: one
// object may answer two instances of the word "target").
func AwakenLandTargets() *game.TargetSpec {
	return TargetPermanent("target land you control", And(Land(), YouControl()))
}

// Awaken is "Awaken N—[cost]" (CR 702.113a): pay `cost` rather than the
// spell's mana cost, and the spell has one more target, "target land
// you control", after its own (CR 702.113b: that target exists only
// when the awaken cost is paid, which is the alternative cost's target
// rewrite). `printed` is the spell's own target statement, the same
// pointer the Spec's Targets holds, or nil for a spell with no target
// (Coastal Discovery). It is copied, never mutated.
//
// The offer declares Purpose.AwakenLand = n for the bot (owner decision
// 6); the spell's own purpose still applies beside it.
func Awaken(n int, cost string, printed *game.TargetSpec) game.AlternativeCost {
	land := AwakenLandTargets()
	var targets *game.TargetSpec
	if printed == nil {
		targets = land
	} else {
		head := *printed
		head.Rest = append([]game.TargetClause(nil), printed.Rest...)
		targets = head.Then(land)
	}
	return game.AlternativeCost{
		Key:      awakenKey,
		Label:    fmt.Sprintf("Awaken %d—%s", n, cost),
		ManaCost: cost,
		Targets:  targets,
		Purpose:  game.Purpose{AwakenLand: n},
	}
}

// AwakenAfter wraps a spell's resolution with its awaken clause: run
// `effect`, then, if the awaken cost was paid, awaken the land the last
// clause named (AwakenIfPaid). Nil `effect` is a spell whose only
// instruction is the awaken. The printed sentence comes first and the
// keyword's second (CR 608.2c).
//
// A spell whose own instruction can pause before it finishes (a counter
// placement waiting on a CR 616 ordering prompt) calls AwakenIfPaid from
// its own continuation instead, as Earthen Arms does.
func AwakenAfter(n int, effect func(*game.StackItem, *Context) error) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if effect != nil {
			if err := effect(item, ctx); err != nil {
				return err
			}
		}
		return AwakenIfPaid(ctx, n)
	}
}

// AwakenIfPaid is the awaken clause on its own: when the spell was cast
// for its awaken cost and its land target is still legal (CR 608.2b),
// put n +1/+1 counters on that land and make it a 0/0 Elemental
// creature with haste (game.AwakenForEffect). A no-op otherwise: the
// spell was cast for its mana cost, or the land left or changed hands.
//
// The land is the pick in the statement's LAST clause, which is the one
// Awaken appends (Register holds every awaken offer to that shape), and
// awaken being paid means that clause was answered.
func AwakenIfPaid(ctx *Context, n int) error {
	if !ctx.PaidAltCost(awakenKey) {
		return nil
	}
	slot := -1
	for _, t := range ctx.Targets() {
		if t.Mode == 0 && t.Slot > slot {
			slot = t.Slot
		}
	}
	if slot < 0 {
		return nil
	}
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ctx.Game.AwakenForEffect(ctx.Controller(), ctx.Source(), t.ID, n)
}

// awakenSpellCardTarget is the spell's OWN target, the pick that answered
// its printed clause (slot 0), when it is a card and still legal (CR
// 608.2b). Read by slot, never as "the first legal target": with awaken
// paid the statement also holds the land, and a printed target that
// left must not hand the instruction to the land.
func awakenSpellCardTarget(ctx *Context) (uuid.UUID, bool) {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return uuid.Nil, false
	}
	return t.ID, true
}

// destroyAwakenSpellTarget is "Destroy target <permanent>" as the printed
// half of an awaken spell (Ruinous Path, Sheer Drop).
func destroyAwakenSpellTarget(_ *game.StackItem, ctx *Context) error {
	if id, ok := awakenSpellCardTarget(ctx); ok {
		return DestroyTarget{Target: id}.Apply(ctx)
	}
	return nil
}

// checkAwaken is Register's guard for an awaken offer: the key and the
// purpose come together, the spell is not modal (none prints awaken),
// and the offer's target statement is the spell's own clauses, in
// order, followed by "target land you control" — so the two cannot
// drift and AwakenIfPaid's "last clause" is the land.
func checkAwaken(spec Spec, ac game.AlternativeCost) {
	isKey := ac.Key == awakenKey
	if !isKey && ac.Purpose.AwakenLand == 0 {
		return
	}
	fail := func(why string) {
		panic(fmt.Sprintf("effects.Register: %q offers %q: %s (ADR 0135 §3)", spec.Name, ac.Key, why))
	}
	if !isKey {
		fail("declares Purpose.AwakenLand on an offer that is not awaken — build it with effects.Awaken")
	}
	if ac.Purpose.AwakenLand <= 0 {
		fail("is an awaken offer with no AwakenLand — build it with effects.Awaken")
	}
	if spec.Modes != nil {
		fail("awaken on a modal spell, which no card prints")
	}
	if ac.ClearsTargets || ac.Targets == nil {
		fail("awaken with no target statement — build it with effects.Awaken")
	}
	want := spec.Targets.ClauseCount()
	if ac.Targets.ClauseCount() != want+1 {
		fail(fmt.Sprintf("awaken's statement has %d clauses, want the spell's %d plus the land", ac.Targets.ClauseCount(), want))
	}
	for i := 0; i < want; i++ {
		a, b := ac.Targets.Clause(i), spec.Targets.Clause(i)
		if a.Label != b.Label || a.Min != b.Min || a.Max != b.Max || a.Distinct != b.Distinct {
			fail(fmt.Sprintf("awaken's clause %d (%q) is not the spell's own (%q) — pass the Spec's Targets to effects.Awaken", i, a.Label, b.Label))
		}
	}
	if last := ac.Targets.Clause(want); last.Label != AwakenLandTargets().Label {
		fail(fmt.Sprintf("awaken's last clause is %q, want %q", last.Label, AwakenLandTargets().Label))
	}
}
