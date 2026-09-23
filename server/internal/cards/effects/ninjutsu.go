package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ninjutsu.go — #1227: ninjutsu (CR 702.49).
//
// "Ninjutsu {1}{U}" is not an alternative way to cast the card and not
// a special action. It is an ACTIVATED ABILITY that functions only
// while the card is in its owner's HAND (CR 702.49a):
//
//	"{1}{U}, Return an unblocked attacker you control to hand: Put
//	 this card onto the battlefield from your hand tapped and
//	 attacking."
//
// Which makes it cycling's sibling — the same CR 602 activation, the
// same stack item, the same cost validation, reaching the hand through
// ActivatedAbilityShape.Zones (ADR 0062 Decision 1) — with a different
// cost component and a different verb on the right of the colon.
// Nothing in the engine knows the word "ninjutsu".
//
// The two halves that did not exist before this file, and both are
// engine shapes rather than card text:
//
//   - the COST is AbilityCost.ReturnToHand (#1213) narrowed to
//     CR 509.1h's "unblocked attacker", which needed the blocked
//     record to be readable from outside internal/game at all
//     (Game.UnblockedAttackerForEffect). The narrowing is a PREDICATE
//     on the clause rather than a bit on the component, so the one
//     candidate walk the client's picker, the legal-move enumerator
//     and the validator all share narrows for all three at once
//     (#544) and no fourth reader can disagree.
//
//   - the EFFECT is an entry that puts a card onto the battlefield
//     ATTACKING (game.ZoneEntryOptions.Attacking, CR 506.3c). Before
//     it, only a minted TOKEN could enter attacking.
//
// And one fact has to cross between them: WHICH attack the ninja
// inherits. CR 702.49a says the same player or planeswalker the
// returned creature was attacking, and that is unrecomputable by the
// time the ability resolves — the return CLEARS Card.AttackingTarget
// and LKI carries no combat state — so the cost's payer reads it
// before the bounce and parks it on the announcement's paid-cost
// record (PaidCost.ReturnedAttacking, ADR 0073's #1227 amendment).
//
// WHAT NINJUTSU DOES NOT NEED. Combat damage needs no change:
// assignAndDealCombatDamageLocked builds its attacker set off
// Card.AttackingTarget and asks the blocked record only for the
// blocked question, so a ninja that arrives after blockers deals its
// damage unblocked. Summoning sickness needs no exemption either: the
// ninja is never DECLARED as an attacker (CR 506.3c), and CR 302.6
// restricts declaring, not attacking.

// UnblockedAttacker matches a creature that is an UNBLOCKED ATTACKER
// right now (CR 509.1h) — ninjutsu's "an unblocked attacker you
// control". "You control" is deliberately NOT part of it: every return
// cost enforces that in the validator, on the field that means it,
// rather than asking a filter that could forget.
//
// The whole of the rule is Game.UnblockedAttackerForEffect's, so the
// predicate cannot drift from the engine's own answer: attacking, the
// cursor at the declare-blockers step or later, no blocked record and
// no blocker staged against it. That function's doc comment carries
// the sandbox caveat that comes with it.
func UnblockedAttacker() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.UnblockedAttackerForEffect(c.InstanceID)
	}
}

// ReturnAnUnblockedAttacker is ninjutsu's cost component —
// ReturnAPermanentToHand narrowed to CR 509.1h's unblocked attacker.
// Exported beside the keyword rather than hidden inside it because the
// clause is a printed one and the next card to print it need not be a
// Ninja.
func ReturnAnUnblockedAttacker() game.AbilityCost {
	return ReturnAPermanentToHand("an unblocked attacker you control", Creature(), UnblockedAttacker())
}

// Ninjutsu is "Ninjutsu <cost>" — CR 702.49a. `cost` is the printed
// mana cost of the ability, "{1}{U}" or "{2}{U}{B}"; the return, the
// hand-only restriction and the whole of the entry come with the
// keyword and are not the card file's to spell out.
//
// Give the ability to the card the way its oracle text reads:
//
//	Activated: []ActivatedAbility{Ninjutsu("{1}{U}")},
//
// There is no "Activate only during the declare blockers step"
// Condition, and adding one would say the same thing twice: a player
// can't begin to activate an ability whose cost they can't pay
// (CR 118.3), and outside that window no creature is an unblocked
// attacker — so the COST is the restriction. The legal-move
// enumerator offers no activation at all when nothing can pay (#544),
// which is the same rule read through the bots.
func Ninjutsu(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Ninjutsu " + cost + " (" + cost + ", Return an unblocked attacker you control to hand: " +
			"Put this card onto the battlefield from your hand tapped and attacking.)",
		Cost:   Plus(ManaCost(cost), ReturnAnUnblockedAttacker()),
		Zones:  []game.ZoneKind{game.ZoneHand},
		Effect: ninjutsuEnter,
	}
}

// ninjutsuEnter is the ability body. The source is its own target in
// everything but name — "this card" — so it comes off the stack item
// rather than off a target clause, which is also why ninjutsu cannot
// fizzle.
//
// Package-level rather than a closure, for the reason every delayed
// trigger body is: nothing may be captured, because undo resolves
// these against a cloned game.
func ninjutsuEnter(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	// CR 608.2a: the card may have left the hand while the ability was
	// on the stack — a discard in response, a Thoughtseize. The ability
	// does as much as it can, which is nothing. Checked rather than
	// assumed, and the OWNER is checked with it: the entry door refuses
	// a card that is not in a hand, but a card now in somebody else's
	// hand is not this ability's card any more (CR 400.7).
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand || z.Owner != item.Controller {
		return nil
	}
	_, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{
		// CR 108.4 made the activator the card's owner, so "under your
		// control" and "under its owner's control" are the same player
		// here — named anyway, because the rule the card prints is the
		// first one.
		Controller: item.Controller,
		Tapped:     true,
		// CR 702.49a: attacking the same player or planeswalker the
		// returned creature was attacking. Read off the announcement's
		// paid-cost record, because the return that paid for this has
		// already cleared the answer off the board.
		Attacking: NewContext(g, item).ReturnedAttacking(),
	})
	return err
}
