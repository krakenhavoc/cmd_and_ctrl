package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_shields.go — constructors for Spec.SpellsCantBeCountered,
// a permanent's printed "<these> spells can't be countered" (ADR 0106
// §4, #1806). The engine half is game/cant_be_countered.go.
//
// One constructor per "whose", named for how the clause reads, and
// the spell filter written in the ordinary CardPredicate vocabulary
// (Creature(), OfColor("G"), Subtype("Sliver"), ManaValueGE(5)),
// ANDed left to right:
//
//	SpellsYouControlCantBeCountered("Spells you control can't be countered.")               // Chimil
//	SpellsYouControlCantBeCountered("Green spells you control can't be countered.", OfColor("G")) // Allosaurus Shepherd
//	SpellsYouCastCantBeCountered("…", ManaValueGE(5))                                       // Thryx
//	AnyPlayersSpellsCantBeCountered("Creature spells can't be countered.", Creature())       // Gaea's Herald
//
// The predicates see the spell as it sits on the stack. The layer
// pass does not recompute a card there, so a type, colour or power is
// the printed one (with a keyword counter's keyword, and X in the mana
// value, as everywhere else); the `caster` a predicate is handed is
// the controller of the permanent with the static.

// SpellsYouControlCantBeCountered is "<these> spells you control can't
// be countered": the spell's current controller is the static's
// controller.
func SpellsYouControlCantBeCountered(label string, preds ...CardPredicate) game.CounterShieldStatic {
	return counterShield(game.CounterShieldYouControl, label, preds)
}

// SpellsYouCastCantBeCountered is "<these> spells you cast … can't be
// countered": the static's controller cast the spell. A copy is not
// cast (CR 707.10) and is never covered.
func SpellsYouCastCantBeCountered(label string, preds ...CardPredicate) game.CounterShieldStatic {
	return counterShield(game.CounterShieldYouCast, label, preds)
}

// AnyPlayersSpellsCantBeCountered is "<these> spells can't be
// countered" with no "you": every player's matching spell.
func AnyPlayersSpellsCantBeCountered(label string, preds ...CardPredicate) game.CounterShieldStatic {
	return counterShield(game.CounterShieldAnyPlayer, label, preds)
}

func counterShield(whose game.CounterShieldWhose, label string, preds []CardPredicate) game.CounterShieldStatic {
	s := game.CounterShieldStatic{Label: label, Whose: whose}
	if len(preds) > 0 {
		match := And(preds...)
		s.Spell = func(g *game.Game, spell game.Card, source *game.Card) bool {
			return match(g, source.Controller, spell)
		}
	}
	return s
}

// --- ADR 0106 PR 3: turn grants, one-use promises and marks -------
//
// The three shapes that are not a printed static, each a primitive a
// card's effect applies (game/counter_shield_grants.go is the engine
// half):
//
//	GrantCounterShield{From: "Veil of Summer", Grant: SpellsCantBeCounteredThisTurn(
//	    "Spells you control can't be countered this turn.", game.CounterShieldYouControl, game.PermissionFilter{})}
//	GrantCounterShield{From: "Insist", Grant: NextSpellYouCastCantBeCountered(
//	    "The next creature spell you cast this turn can't be countered.", game.PermissionFilter{CreatureOnly: true})}
//	MarkTargetSpellsCantBeCountered{From: "Vexing Shusher", Label: "Target spell can't be countered."}
//
// The filter is a game.PermissionFilter, not a CardPredicate, because
// a grant is STORED on the player (and so in a restore point) and a
// predicate is a closure that could not be.

// SpellsCantBeCounteredThisTurn is "<these> spells you control / you
// cast can't be countered this turn": read at the counter gate for the
// rest of the turn, so it covers spells cast after it resolved
// (CR 611.2c) and ends at cleanup (CR 514.2).
func SpellsCantBeCounteredThisTurn(text string, whose game.CounterShieldWhose, filter game.PermissionFilter) game.CounterShieldGrant {
	return game.CounterShieldGrant{Active: true, Whose: whose, Filter: filter, Text: text}
}

// NextSpellYouCastCantBeCountered is "the next <these> spell you cast
// this turn can't be countered": a one-use promise, spent by the first
// matching spell its player casts (CR 601.2i), or ended at cleanup.
func NextSpellYouCastCantBeCountered(text string, filter game.PermissionFilter) game.CounterShieldGrant {
	return game.CounterShieldGrant{Active: true, Whose: game.CounterShieldYouCast, NextOnly: true, Filter: filter, Text: text}
}

// GrantCounterShield gives Player (zero: the controller of the spell
// or ability) a turn grant or a promise. From is the source card's
// name, which the player panel shows beside the clause. ExceptThis
// excepts the spell resolving it, for "OTHER spells you control" (the
// Determined half of Bound // Determined), by object, so the same card
// cast again later is a new object and is covered (CR 400.7).
type GrantCounterShield struct {
	Player     uuid.UUID
	From       string
	Grant      game.CounterShieldGrant
	ExceptThis bool
}

func (a GrantCounterShield) Apply(ctx *Context) error {
	player := a.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	grant := a.Grant
	if a.ExceptThis {
		ref := game.ObjectRef{ID: ctx.Source()}
		if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
			ref.Epoch = c.ObjectEpoch
		}
		grant.Except = ref
	}
	ctx.Game.GrantCounterShieldForEffect(player, grant, a.From, ctx.Source())
	return nil
}

// MarkTargetSpellsCantBeCountered is "Target spell can't be countered"
// (Vexing Shusher): every target that is still a legal spell on the
// stack is marked, for as long as that object stays there (CR 400.7).
// A target that left in response is skipped (CR 608.2b).
type MarkTargetSpellsCantBeCountered struct {
	From  string
	Label string
}

func (a MarkTargetSpellsCantBeCountered) Apply(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		ctx.Game.MarkSpellCantBeCounteredForEffect(t.ID, game.CounterShieldMark{
			Source: ctx.Source(), SourceName: a.From, Label: a.Label,
		})
	}
	return nil
}
