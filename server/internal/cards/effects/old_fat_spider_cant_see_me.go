package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old Fat Spider Can't See Me — Enchantment — Saga {2}{U}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Target creature you control gains hexproof for as long as this
//	     Saga remains on the battlefield.
//	 II — Prevent all damage that would be dealt by up to one target
//	      creature for as long as this Saga remains on the battlefield.
//	 III, IV — Draw a card."
//
// ADR 0108 amendment 2026-10-07 (#2027): both chapters last "for as long
// as this Saga remains on the battlefield" (CR 611.2b). The hexproof is a
// layer-6 grant on the creature as it is now (CR 611.2c), the prevention
// is a preventFromSource shield pinned to the chosen creature
// (ShieldWhileSourceRemains). Both end the moment the Saga leaves, and
// neither starts at all if the Saga is no longer the object the chapter
// came from (CR 611.2b, CR 400.7) — a Saga that was flickered while the
// chapter waited. Chapter II's "up to one" with no target does nothing.
//
// No simplifications.
func init() {
	hexproof := ChapterTriggerTargeting(1, "Old Fat Spider Can't See Me — I: target creature you control gains hexproof",
		TargetCreature("target creature you control", YouControl()), oldFatSpiderHexproof)
	prevent := ChapterTriggerTargeting(2, "Old Fat Spider Can't See Me — II: prevent all damage dealt by up to one target creature",
		TargetCreature("up to one target creature").WithCount(0, 1), oldFatSpiderPrevent)
	Register(Spec{
		OracleID:     "e4402d21-bb7f-4777-bc8d-b24bf2faccb0",
		Name:         "Old Fat Spider Can't See Me",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			hexproof,
			prevent,
			ChapterTrigger(3, "Old Fat Spider Can't See Me — III: draw a card", drawOneChapter),
			ChapterTrigger(4, "Old Fat Spider Can't See Me — IV: draw a card", drawOneChapter),
		},
	})
}

// oldFatSpiderHexproof is chapter I: the target creature gains hexproof
// for as long as the Saga that made the ability remains on the battlefield.
func oldFatSpiderHexproof(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	d, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddKeywordsMod("hexproof")},
			Duration: d,
			Label:    "Old Fat Spider Can't See Me — hexproof while the Saga remains",
		}.Apply(ctx)
	}
	return nil
}

// oldFatSpiderPrevent is chapter II: all damage the chosen creature would
// deal is prevented for as long as the Saga remains.
func oldFatSpiderPrevent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return PreventDamageFromSource{
			From: t.ID, Protect: ShieldAnything, Lasts: ShieldWhileSourceRemains,
			Label: "Old Fat Spider Can't See Me — prevent all damage it would deal while the Saga remains",
		}.Apply(ctx)
	}
	return nil
}

// drawOneChapter is a chapter that reads "Draw a card."
func drawOneChapter(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
