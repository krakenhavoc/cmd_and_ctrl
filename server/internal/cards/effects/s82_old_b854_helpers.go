package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b854_helpers.go — Fallen Empires, keywords, removal and combat
// tricks (S82 OLD b854, #5637).

// S82OldB854BlocksOrIsBlockedBySubtype is the trigger condition of "Whenever
// this creature blocks or becomes blocked by one or more <subtype> creatures"
// (Dwarven Soldier's Orcs). It is one trigger per block, not per pair (CR
// 509.3a, 509.3c): the creature's first EventBlock when at least one attacker it
// blocks has the subtype, or its one EventBecomesBlocked when at least one of
// its blockers has it.
func S82OldB854BlocksOrIsBlockedBySubtype(subtype string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	match := func(c *game.Card) bool { return c.IsCreature() && c.HasSubtype(subtype) }
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		switch ev.Kind {
		case game.EventBlock:
			if ev.CardID != source.InstanceID || ev.Amount > 1 {
				return false
			}
			for _, a := range source.BlockedAttackers() {
				if c, ok := g.LookupCardForEffect(a); ok && match(&c) {
					return true
				}
			}
		case game.EventBecomesBlocked:
			if ev.CardID != source.InstanceID {
				return false
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsBlockingAttacker(source.InstanceID) && match(&c) {
					return true
				}
			}
		}
		return false
	}
}

// S82OldB854TapNOthers is "Tap N untapped <permanents> you control" as an
// activated ability's cost beside the source's own {T}: the source never pays
// it with itself (CR 118.3), and like every tap-others cost it is not the {T}
// symbol, so a creature that arrived this turn may pay it (CR 302.6).
func S82OldB854TapNOthers(n int, label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:         n,
		Filter:        TargetPermanent(label, preds...),
		ExcludeSource: true,
		Label:         label,
	}}
}

// S82OldB854FirstStrikeOnBlock is Goblin Flotilla's delayed trigger: for the
// rest of this combat, whenever the Flotilla blocks or becomes blocked by a
// creature, that creature gains first strike until end of turn.
var (
	s82OldB854FlotillaCondition = game.DelayedEachCondition("combat/flotilla-blocks-or-is-blocked", s82OldB854FlotillaOther)
	s82OldB854FlotillaBody      = game.SimpleDelayedBody("combat/flotilla-first-strike", s82OldB854FlotillaGrant)
)

// s82OldB854FlotillaOther is the "that creature" of one EventBlock pair: the
// attacker when the Flotilla blocks, the blocker when it is blocked. The event
// names the blocker in CardID and the attacker in Target.
func s82OldB854FlotillaOther(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBlock || p.Object.ID == uuid.Nil {
		return nil
	}
	flotilla, ok := BattlefieldCardForEffect(g, p.Object.ID)
	if !ok || flotilla.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	switch p.Object.ID {
	case ev.CardID:
		return []uuid.UUID{ev.Target}
	case ev.Target:
		return []uuid.UUID{ev.CardID}
	}
	return nil
}

func s82OldB854FlotillaGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	it := FirstPayloadCard(ctx)
	if it == uuid.Nil || !OnBattlefield(g, it) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: it, Keywords: []string{"first strike"}, Label: "Goblin Flotilla — first strike until end of turn"}.Apply(ctx)
}

// S82OldB854FlotillaNotPaid is the decline of "unless you pay {R}": the delayed
// trigger is created, and lasts for this combat only (an additional combat is
// another one, CR 511.3).
var S82OldB854FlotillaNotPaid = PayUnlessElse("pay-unless/goblin-flotilla-first-strike", func(ctx *Context, _ game.PayUnlessAnswer) error {
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok || !OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	ref, ok := ctx.Game.PermanentRefForEffect(ctx.Source())
	if !ok {
		return nil
	}
	return DelayedOnEvent{
		Label:      "Goblin Flotilla — that creature gains first strike until end of turn",
		On:         []game.EventKind{game.EventBlock},
		Condition:  s82OldB854FlotillaCondition,
		CondParams: game.EffectParams{Object: ref},
		Body:       s82OldB854FlotillaBody,
		Repeats:    true,
		Duration:   &d,
	}.Apply(ctx)
})

// S82OldB854HasSubtype reports whether a permanent, as read, has a creature type.
func S82OldB854HasSubtype(info game.PermanentInfo, subtype string) bool {
	for _, s := range info.Characteristic.Subtypes {
		if s == subtype {
			return true
		}
	}
	return info.Characteristic.AllCreatureTypes
}
