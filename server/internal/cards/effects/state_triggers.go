package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_triggers.go — the constructors for CR 603.8 state triggers
// (ADR 0107 §1, #1858): "When you control no Islands, sacrifice this
// creature", "When there are no creatures on the battlefield, sacrifice
// this enchantment", "When there are five or more plot counters on this
// enchantment, sacrifice it".
//
// A state trigger watches no event. Its row declares a condition over
// the board (game.TriggeredAbility.State) and its Effect, and the engine
// asks the condition after every event and in each pass of the CR 704.3
// loop, latching the ability while an item of it is waiting or on the
// stack (game/state_triggers.go). A card file names the condition and
// the effect, and nothing else.
//
// Never approximate one with an event trigger (watching a land leave, a
// counter being removed): that triggers once per event instead of once
// per state, and misses every way the state can arise that the card file
// did not think of.

// StateCondition is the TriggeredAbility.State signature: is the game
// in the printed state, for this permanent and its controller? A pure
// read of the board, under the game lock.
type StateCondition = func(g *game.Game, source *game.Card, controller uuid.UUID) bool

// WhenState is the general state-trigger constructor: the ability
// triggers whenever `cond` holds and it is not already waiting or on the
// stack, and resolves as `label` with `effect`. Reach for it when the
// printed condition has no named helper below.
func WhenState(label string, cond StateCondition, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{State: cond, Key: label, Effect: effect}
}

// WhenYouControlNo is "When you control no <permanents matching q>":
// Barbarian Outcast's Swamps, Covetous Dragon's artifacts, Serendib
// Djinn's lands, the serpents' Islands. The query is ADR 0107's shared
// game.PermanentQuery, the one the serpents' attack clause reads
// (QuerySubtype("Island")), so the two halves of a Sea Serpent ask one
// question the same way.
func WhenYouControlNo(q game.PermanentQuery, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
		return !g.ControlsMatchingForEffect(controller, uuid.Nil, q)
	}, effect)
}

// WhenYouControlNoOther is "When you control no OTHER <permanents
// matching q>" — the source itself does not count: Emperor Crocodile's
// creatures, Synod Centurion's artifacts.
func WhenYouControlNoOther(q game.PermanentQuery, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(g *game.Game, source *game.Card, controller uuid.UUID) bool {
		return !g.ControlsMatchingForEffect(controller, source.InstanceID, q)
	}, effect)
}

// WhenThereAreNo is "When there are no <permanents matching q> on the
// battlefield" — anyone's: Task Mage Assembly's creatures, Mana
// Vortex's lands.
func WhenThereAreNo(q game.PermanentQuery, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(g *game.Game, _ *game.Card, _ uuid.UUID) bool {
		return !g.AnyPermanentMatchingForEffect(q)
	}, effect)
}

// WhenYouControlAtLeast is "When you control N or more <permanents
// matching q>": Endrek Sahr's seven Thrulls.
func WhenYouControlAtLeast(n int, q game.PermanentQuery, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
		return g.CountControlledMatchingForEffect(controller, q) >= n
	}, effect)
}

// QueryType is "a permanent of this card type": an artifact, a land, a
// creature. The card-type sibling of QuerySubtype.
func QueryType(cardType string) game.PermanentQuery {
	return game.PermanentQuery{Types: []string{cardType}}
}

// WhenThisHasAtLeast is "When there are N or more <kind> counters on
// this permanent" (Deadly Designs, Mazemind Tome, Nine Lives) and its
// "this has N or more <kind> counters on it" wording (Darksteel Reactor,
// Plague Boiler).
func WhenThisHasAtLeast(kind string, n int, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
		return source.Counters[kind] >= n
	}, effect)
}

// WhenThisHasNo is "When this has no <kind> counters on it": Dark
// Depths' ice counters, Afiya Grove's +1/+1 counters.
func WhenThisHasNo(kind string, label string, effect Effect) game.TriggeredAbility {
	return WhenState(label, func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
		return source.Counters[kind] == 0
	}, effect)
}

// controlsAnyCard reports whether `controller` controls a permanent pred
// accepts, for a state whose permanents no PermanentQuery describes
// (Endangered Armodon's "a creature with toughness 2 or less"). pred is
// asked with the controller as its "caster", which is what every
// relative predicate reads.
func controlsAnyCard(g *game.Game, controller uuid.UUID, pred CardPredicate) bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && pred(g, controller, *c) {
			return true
		}
	}
	return false
}

// SacrificeThisThen is "sacrifice it. If you do, <then>": the state
// trigger's sacrifice of its own source, with the follow-up run only when
// the permanent really left (SacrificePermanent.Then). A source that is
// no longer on the battlefield, or is a new object (CR 400.7), is not
// sacrificed, and `then` does not run.
func SacrificeThisThen(then func(ctx *Context) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return SacrificePermanent{Target: item.SourceCardID, Then: func(ctx *Context, sacrificed bool) error {
			if !sacrificed || then == nil {
				return nil
			}
			return then(ctx)
		}}.Apply(NewContext(g, item))
	}
}

// ExileThisThen is "exile it. If you do, <then>" for a state trigger's
// own source (Mazemind Tome); a plain "exile it" passes nil.
func ExileThisThen(then func(ctx *Context) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return ExileTarget{Target: item.SourceCardID, Then: func(ctx *Context, exiled bool) error {
			if !exiled || then == nil {
				return nil
			}
			return then(ctx)
		}}.Apply(NewContext(g, item))
	}
}

// checkStateTrigger is Register's part of the contract: a state trigger
// watches nothing, declares its Effect (so its stack item is keyed and a
// table with one waiting is a restore point, ADR 0041 P9), and lives on
// the battlefield, the only zone the engine asks.
func checkStateTrigger(name string, t game.TriggeredAbility) {
	if t.State == nil {
		return
	}
	switch {
	case len(t.Watches) > 0:
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger that also watches events — a CR 603.8 state trigger watches no event", name))
	case t.Effect == nil:
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger with no Effect — declare it on the row (ADR 0041 P9)", name))
	case t.Build != nil:
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger with a Build — the engine builds a state trigger's item from its Key and Effect", name))
	case t.Key == "":
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger with no Key — the label names the ability on the stack and in the CR 603.8 latch", name))
	case len(t.Zones) > 0:
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger outside the battlefield — the engine asks state triggers on permanents only", name))
	case t.OncePerBatch || t.FromStack:
		panic(fmt.Sprintf("effects.Register: %q declares a state trigger with an event-batch or cast-from-stack flag — neither means anything for a state", name))
	}
}
