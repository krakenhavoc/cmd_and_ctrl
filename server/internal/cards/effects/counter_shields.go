package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
