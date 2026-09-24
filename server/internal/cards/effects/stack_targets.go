package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_targets.go — "target spell or ability" (CR 115.4), the clause
// half of #1211.
//
// The stack is the one zone that holds two KINDS of object: spell
// cards, which live in `game.Game.Stack`, and activated / triggered
// ability items, which live in `Game.StackMeta` with a synthetic id
// and no card anywhere. `TargetSpell` (targets.go) walks the zone and
// sees only the first; `AbilityOnStack` (ability_copy.go, #1223) reads
// `StackMeta` and sees only the second. Everything printed "target
// spell OR ability" needs both at once, and that is the whole of this
// file.
//
// One id space, one ref kind: a picked ability rides an ordinary
// `game.TargetRef{Kind: game.TargetCard}` carrying the STACK ITEM's
// id — see `game.TargetSpec.Abilities` for why, and
// docs/decisions/0019-structured-targeting.md's 2026-09-23 amendment
// for the decision.
//
// **A mana ability is unreachable for free.** CR 605.3b keeps it off
// the stack, so there is no item to enumerate and the printed "(Mana
// abilities can't be targeted.)" on Stifle, Disallow and Voidslime
// needs no predicate of its own.

// StackItemPredicate narrows a clause over the STACK, where a spell
// and an ability are the same kind of thing: a stack item, with the
// same announcement record behind it (targets, modes, X, what was
// paid).
//
// It is the general name for the function type `AbilityPredicate`
// already is — the two are the same Go type and are interchangeable
// at every call site. Use this one when the clause admits spells too.
//
// Runs under g.mu like every other target predicate: read-only
// *ForEffect accessors are fine, public locking mutators are not.
type StackItemPredicate = func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool

// TargetSpellOrAbility — "target spell or ability" (CR 115.4).
//
// The clause Deflecting Swat, Bolt Bend, Disallow, Voidslime and
// Tale's End print. One spec, two enumerations: `Zones: {ZoneStack}`
// finds the spell cards, `Abilities: true` finds the ability items,
// and both answer the same `TargetCard` ref.
//
// The predicates narrow BOTH halves, because the printed restrictions
// on this clause are facts about the ANNOUNCEMENT rather than about a
// card — Bolt Bend's "with a single target" counts slots on the item,
// and a spell and an ability both have one. A restriction that really
// is about one half only is written with `ASpellItem` /
// `AnAbilityItem` inside `AnyStackItem`, which is how Tale's End says
// "activated ability, triggered ability, or LEGENDARY spell".
//
// The spell half refuses a card on the stack with no `StackMeta`
// entry. That is strictly narrower than `TargetSpell`'s bare zone walk
// and deliberately so: the only object in that state is a spell that
// is MID-RESOLUTION (the meta is deleted before the card is routed
// away), and nothing may target one.
func TargetSpellOrAbility(label string, preds ...StackItemPredicate) *game.TargetSpec {
	pred := AllAbilities(preds...)
	return &game.TargetSpec{
		Mode:  "stack_item",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneStack},
		CardOK: func(g *game.Game, chooser uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			item := g.StackItemForEffect(c.InstanceID)
			return item != nil && item.Kind == game.StackItemSpell && pred(g, chooser, item)
		},
		Abilities: true,
		AbilityOK: func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
			return pred(g, chooser, item)
		},
		Min: 1, Max: 1,
	}
}

// AnyStackItem composes stack-item predicates with OR — the
// `AllAbilities` (AND) twin, and the shape a "this, that, or the
// other" clause needs:
//
//	AnyStackItem(AnAbilityItem(), ASpellItem(Legendary()))
//	  → "target activated ability, triggered ability, or legendary spell"
//
// With no predicates it admits nothing, which is the honest identity
// for an OR and never what a card file means to write — the empty
// clause is `TargetSpellOrAbility(label)` with no predicates at all.
func AnyStackItem(preds ...StackItemPredicate) StackItemPredicate {
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		for _, p := range preds {
			if p(g, chooser, item) {
				return true
			}
		}
		return false
	}
}

// AnAbilityItem passes for an activated or triggered ability item and
// fails for a spell — the "activated ability, triggered ability"
// two-thirds of Tale's End, and the way to say "the ability half" of
// a clause that admits both.
//
// There is no third kind: `StackItemKind` is spell / activated /
// triggered, and a mana ability never reaches the stack at all.
func AnAbilityItem() StackItemPredicate {
	return func(_ *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		return item.Kind != game.StackItemSpell
	}
}

// ASpellItem passes for a SPELL on the stack whose card satisfies
// every `preds` — the "or legendary spell" third of Tale's End,
// written as the ordinary card vocabulary applied to the item's card.
//
// The card is read as it is NOW rather than as it was announced, which
// is the rule for every target predicate: a clause is checked at
// announce (CR 601.2c) and again at resolution (CR 608.2b), and a
// spell whose characteristics changed in between is judged on what it
// is at each check.
func ASpellItem(preds ...CardPredicate) StackItemPredicate {
	pred := And(preds...)
	return func(g *game.Game, chooser uuid.UUID, item *game.StackItem) bool {
		if item.Kind != game.StackItemSpell {
			return false
		}
		c, ok := g.LookupCardForEffect(item.ID)
		return ok && pred(g, chooser, c)
	}
}

// ItemHasASingleTarget is "with a single target" over a stack ITEM —
// Bolt Bend's printed clause, which restricts a spell and an ability
// with the same words and must therefore restrict them with the same
// function.
//
// The `CardPredicate` twin `HasASingleTarget` (retarget.go, #1196) is
// still what "target SPELL with a single target" is written with
// (Misdirection, Ricochet Trap, Imp's Mischief). Both read
// `StackItemTargetCountForEffect`, so the two cannot drift: it counts
// SLOTS rather than distinct objects, and a spell that chose one
// creature for two instances of the word "target" (CR 115.3's
// AllowSame) has two targets and is not a single-target spell.
func ItemHasASingleTarget() StackItemPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		return g.StackItemTargetCountForEffect(item.ID) == 1
	}
}
