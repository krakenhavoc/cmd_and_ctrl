package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// replicate.go — CR 702.56a's keyword, ADR 0129 §5 (PR 4):
//
//	"Replicate [cost]" means "As an additional cost to cast this spell,
//	 you may pay [cost] any number of times" and "When you cast this
//	 spell, if a replicate cost was paid for it, copy it for each time
//	 its replicate cost was paid. If the spell has any targets, you may
//	 choose new targets for any of the copies."
//
// Two halves, each an existing seam:
//
//   - The cost is an optional cost (ADR 0073) keyed game.ReplicateKey
//     and paid any number of times, the multikicker shape. Reiterating
//     Bolt's "Pay {E}{E}{E}" is its AdditionalCost.Energy, so the cast
//     path sums it per payment, checks it against the caster's energy
//     (CR 118.3) and pays it through the one energy payer.
//   - The trigger is storm's (effects/storm.go): FromStack, watching
//     the spell's own EventCast, copying through CopySpell with new
//     targets offered. The count is the paid record's, read when the
//     trigger is built (CR 702.56a's "if a replicate cost was paid" is
//     an intervening if, CR 603.4, so a spell with none paid triggers
//     nothing) and carried on the item's XValue, so a replicate trigger
//     waiting on the stack is a restore point and a spell countered in
//     response is still copied from last-known information (CR 608.2h).
//
// A card declares both, and Register refuses one without the other:
//
//	OptionalCosts: []game.AdditionalCost{ReplicatePayEnergy(3, 10)},
//	Triggered:     []game.TriggeredAbility{Replicate()},

// replicateKey is the Key of every replicate trigger row — the name a
// restored item checks its row by (ADR 0041 P9).
const replicateKey = "replicate"

// KeywordReplicate is the printed keyword the trigger row carries.
const KeywordReplicate = "replicate"

// ReplicatePayEnergy is "Replicate—Pay N {E}" (Reiterating Bolt's
// three): the optional cost, paid any number of times. `max` is the
// engine's cap on one announcement, multikicker's reason: replicate is
// unbounded in paper, an announcement is not, and energy bounds it far
// below any cap a card file picks (CR 118.3). Say so in the card's
// comment.
func ReplicatePayEnergy(n, max int) game.AdditionalCost {
	if max < 1 {
		max = 1
	}
	return game.AdditionalCost{
		Optional: true,
		Key:      game.ReplicateKey,
		Energy:   n,
		Repeat:   max,
		Label:    "Replicate—Pay " + EnergySymbols(n),
	}
}

// Replicate is the keyword's trigger: "When you cast this spell, if a
// replicate cost was paid for it, copy it for each time its replicate
// cost was paid. If the spell has any targets, you may choose new
// targets for any of the copies." (CR 702.56a).
func Replicate() game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword:   KeywordReplicate,
		FromStack: true,
		Key:       replicateKey,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.CardID == source.InstanceID && replicatePayments(g, source) > 0
		},
		// A fill-in Build (ADR 0041 P9): the caster, the spell and the
		// count are facts of the cast; the effect is the row's.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			return &game.StackItem{
				Kind:         game.StackItemTriggered,
				Controller:   ev.Actor,
				Owner:        ev.Actor,
				SourceCardID: source.InstanceID,
				Label:        source.Name + " — replicate",
				XValue:       replicatePayments(g, source),
			}
		},
		Effect: replicateEffect,
	}
}

// replicatePayments is how many times the spell's replicate cost was
// paid, read off the record its cast put on the stack (CR 702.56a).
func replicatePayments(g *game.Game, spell *game.Card) int {
	if g == nil || spell == nil {
		return 0
	}
	return game.OptionalCostTimesPaid(*spell, g.StackItemPaidForEffect(spell.InstanceID).OptionalCosts, game.ReplicateKey)
}

// replicateEffect copies the spell once per payment. CopySpell offers
// each copy new targets (CR 707.10c) only when the spell chose any, so a
// spell with no targets asks nothing.
func replicateEffect(g *game.Game, item *game.StackItem) error {
	if item.XValue <= 0 {
		return nil
	}
	return CopySpell{
		StackID:          item.SourceCardID,
		Controller:       item.Controller,
		Count:            item.XValue,
		ChooseNewTargets: true,
		// CR 608.2h, storm's reason (#1255): the trigger names the spell
		// and does not target it, so a spell countered in response is
		// still copied.
		FromLastKnown: true,
	}.Apply(NewContext(g, item))
}

// checkReplicate is Register's guard for replicate and for energy in an
// additional cost (ADR 0129 §5). Energy is a component of a replicate
// cost and of nothing else: no Commander-legal card prints "As an
// additional cost to cast this spell, pay {E}", and a mandatory or
// branched energy cost has no wire or enumerator shape yet. A replicate
// cost needs its trigger and the trigger its cost, or the card pays for
// copies it never makes (or makes copies of nothing).
func checkReplicate(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil {
		if ac.Energy != 0 {
			panic(fmt.Sprintf("effects.Register: %q declares energy on its mandatory additional cost — energy rides a replicate cost only (ADR 0129 §5)", spec.Name))
		}
		for _, b := range ac.Either {
			if b.Energy != 0 {
				panic(fmt.Sprintf("effects.Register: %q declares energy on an either/or branch — energy rides a replicate cost only (ADR 0129 §5)", spec.Name))
			}
		}
	}
	costs := 0
	for _, oc := range spec.OptionalCosts {
		if oc.Energy < 0 {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q has a negative Energy", spec.Name, oc.Key))
		}
		if oc.Energy != 0 && oc.Key != game.ReplicateKey {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q pays energy — only a replicate cost does (ADR 0129 §5)", spec.Name, oc.Key))
		}
		if oc.Key == game.ReplicateKey {
			costs++
		}
	}
	triggers := 0
	for _, t := range spec.Triggered {
		if t.Key == replicateKey {
			triggers++
		}
	}
	if costs > 1 {
		panic(fmt.Sprintf("effects.Register: %q declares %d replicate costs — CR 702.56b's several instances are not modelled", spec.Name, costs))
	}
	if costs != triggers {
		panic(fmt.Sprintf("effects.Register: %q declares %d replicate cost(s) and %d replicate trigger(s) — declare ReplicatePayEnergy with Replicate()", spec.Name, costs, triggers))
	}
}
