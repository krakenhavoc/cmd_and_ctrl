package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_cost.go — #6190: the card side of a NON-MANA cost to attack or
// block (CR 508.1h, 509.1d). The engine half is game/combat_cost.go,
// which pays it through the very cost components an activated ability's
// cost uses, so everything here is written in the vocabulary a card
// file already writes a cost in (SacrificeN, TapAnUntapped,
// ReturnAPermanentToHand).
//
//	Exalted Dragon:
//	    AttackTaxes: []game.AttackTax{
//	        CantAttackUnlessYouSacrifice("a land", Land()),
//	    },
//
//	Leviathan:
//	    CantAttackUnlessYouSacrificeN(2, "two Islands", HasSubtype("Island")),
//
//	Flooded Woodlands ("Green creatures can't attack unless their
//	controller sacrifices a land of their choice for each green creature
//	they control that's attacking"):
//	    CreaturesCantAttackUnlessControllerSacrifices(
//	        func(q game.AttackTaxQuery) bool { return q.Attacker.HasColor("G") },
//	        1, "a land", Land()),
//
//	Floodtide Serpent:
//	    CantAttackUnlessYouReturn("an enchantment you control", Enchantment()),
//
//	Hollow Warrior ("can't attack or block unless you tap an untapped
//	creature you control not declared as an attacking or blocking
//	creature this combat"):
//	    AttackTaxes: []game.AttackTax{CantAttackUnlessYouTap(hollowWarriorClause, NotInCombatThisCombat(), Creature(), Untapped())},
//	    BlockTaxes:  []game.BlockTax{CantBlockUnlessYouTap(hollowWarriorClause, NotInCombatThisCombat(), Creature(), Untapped())},
//
// Two things these constructors settle that a card file must not have
// to think about:
//
//   - the cost is the ATTACKING creature's controller's, whatever they
//     attack (game.AttackTax.OnAttacker), because "this creature can't
//     attack unless you sacrifice a land" says nothing about who is
//     attacked. The attack tax's structural "you" — the defending
//     player — is not involved;
//   - the clause binds the creature that PRINTS it unless the card says
//     otherwise, which is SelfAttacker below; Flooded Woodlands and
//     Reclamation, which bind every player's green or black attackers,
//     pass their own predicate instead.

// SelfAttacker is the AppliesTo for "THIS creature can't attack unless
// …": the taxed creature is the permanent that prints the clause.
func SelfAttacker(q game.AttackTaxQuery) bool {
	return q.Attacker.InstanceID == q.Source.InstanceID
}

// CantAttackUnlessYouSacrifice is "This creature can't attack unless you
// sacrifice <label>" (Exalted Dragon's land). The sacrifice is paid as
// attackers are declared (CR 508.1h) and goes through the engine's own
// sacrifice payer, so dies-triggers and "whenever you sacrifice" payoffs
// fire exactly as they do for an ability's cost.
func CantAttackUnlessYouSacrifice(label string, preds ...CardPredicate) game.AttackTax {
	return CantAttackUnlessYouSacrificeN(1, label, preds...)
}

// CantAttackUnlessYouSacrificeN is the same clause with a count —
// Leviathan's "unless you sacrifice two Islands".
func CantAttackUnlessYouSacrificeN(n int, label string, preds ...CardPredicate) game.AttackTax {
	return selfAttackCost(label, game.CombatCost{
		Sacrifice: SacrificeSpec(label, preds...).WithCount(n, n),
		Label:     label,
	})
}

// CantAttackUnlessYouTap is "This creature can't attack unless you tap
// <label>" (Hollow Warrior's untapped creature). TapOthersCost, not the
// {T} symbol: a creature that arrived this turn may be tapped to pay it
// (CR 302.6, 602.5a).
func CantAttackUnlessYouTap(label string, preds ...CardPredicate) game.AttackTax {
	return selfAttackCost(label, game.CombatCost{
		Tap:   &game.TapOthersCost{Count: 1, Filter: TargetPermanent(label, preds...), Label: label},
		Label: label,
	})
}

// CantAttackUnlessYouReturn is "This creature can't attack unless you
// return <label> to its owner's hand" (Floodtide Serpent's enchantment).
// To its OWNER's hand (CR 109.5), which the shared payer gets right.
func CantAttackUnlessYouReturn(label string, preds ...CardPredicate) game.AttackTax {
	return selfAttackCost(label, game.CombatCost{
		Return: &game.ReturnToHandCost{Count: 1, Filter: TargetPermanent(label, preds...), Label: label},
		Label:  label,
	})
}

// CreaturesCantAttackUnlessControllerSacrifices is the clause that binds
// creatures OTHER than its source: "Green creatures can't attack unless
// their controller sacrifices a land of their choice for each green
// creature they control that's attacking" (Flooded Woodlands), and
// Reclamation's black half.
//
// `applies` chooses the attackers it binds, read at declaration off
// q.Attacker's effective characteristics. The count is PER TAXED
// ATTACKER, which is what "for each ... that's attacking" means: the
// engine sums the clause once per attacking creature it binds, so three
// green attackers owe three lands.
func CreaturesCantAttackUnlessControllerSacrifices(applies func(q game.AttackTaxQuery) bool, n int, label string, preds ...CardPredicate) game.AttackTax {
	t := CantAttackUnlessYouSacrificeN(n, label, preds...)
	t.AppliesTo = applies
	return t
}

// selfAttackCost is the shared body: a cost the attacking creature's
// controller owes whatever they attack, bound to the permanent that
// prints it.
func selfAttackCost(label string, cost game.CombatCost) game.AttackTax {
	return game.AttackTax{
		Label:      label,
		OnAttacker: true,
		AppliesTo:  SelfAttacker,
		NonMana:    cost,
	}
}

// CantBlockUnlessYouTap is "This creature can't block unless you tap
// <label>" (Hollow Warrior's block half). A block tax prints no "you",
// so the clause is narrowed to the permanent that prints it with
// game.BlockedBySelf.
func CantBlockUnlessYouTap(label string, preds ...CardPredicate) game.BlockTax {
	return game.BlockTax{
		Label:     label,
		AppliesTo: game.BlockedBySelf,
		NonMana: game.CombatCost{
			Tap:   &game.TapOthersCost{Count: 1, Filter: TargetPermanent(label, preds...), Label: label},
			Label: label,
		},
	}
}

// CantBlockUnlessYouSacrifice is the block twin of
// CantAttackUnlessYouSacrifice, for the day a card prints it. Kept
// beside its sibling rather than invented later, because the engine
// pays both halves through one payer and a card file should not have to
// discover that the block half exists.
func CantBlockUnlessYouSacrifice(label string, preds ...CardPredicate) game.BlockTax {
	return game.BlockTax{
		Label:     label,
		AppliesTo: game.BlockedBySelf,
		NonMana: game.CombatCost{
			Sacrifice: SacrificeSpec(label, preds...).WithCount(1, 1),
			Label:     label,
		},
	}
}

// NotInCombatThisCombat is the predicate for "not declared as an
// attacking or blocking creature this combat" (Hollow Warrior): a
// creature that is attacking or blocking right now, or that was
// DECLARED as an attacker or a blocker earlier in this combat phase and
// has since left combat, cannot pay.
//
// The declaration is history, so it is read off the turn's records
// (game.DeclaredAsAttackerThisCombat / DeclaredAsBlockerThisCombat,
// keyed by object per CR 400.7) rather than off the creature's current
// combat state alone — a creature removed from combat was still
// declared.
func NotInCombatThisCombat() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			return false
		}
		return !g.DeclaredAsAttackerThisCombat(&c) && !g.DeclaredAsBlockerThisCombat(&c)
	}
}
