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
// no blocker staged against it — and, since #1279, a defending player
// who has FINISHED declaring blockers, so an attacker whose defender is
// still deciding pays nothing.
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

// CommanderNinjutsu is "Commander ninjutsu <cost>" — CR 702.49c:
//
//	"[cost], Return an unblocked attacker you control to hand: Put
//	 this card onto the battlefield from your hand or the command zone
//	 tapped and attacking."
//
// Ninjutsu with a second zone, and nothing else (#1278). The cost, the
// timing-is-the-cost posture and the attacking entry are Ninjutsu's
// unchanged; the ability also functions from the command zone
// (CR 113.6), and the effect body puts the card from whichever of the
// two zones it is actually in when the ability resolves.
//
// What the command zone brings is bookkeeping, and all of it is a
// negative the engine already answers:
//
//   - No commander tax. CR 903.8 taxes CASTING the commander, and
//     this puts it; Player.CommanderCasts is bumped by the cast path
//     alone, so it neither charges the tax nor adds to the next one.
//   - The designation stays. Card.IsCommander rides the card out of
//     the command zone, so the permanent deals commander damage
//     (CR 903.10a) and, when it leaves the battlefield, the shared
//     exit primitive offers CR 903.9's "command zone instead" as it
//     would for a cast commander.
//   - Activating it is not something only the ability's owner can
//     see: the command zone is public, but the ability ROW rides
//     zone_abilities, which the wire gives only to the card's owner
//     (the CR 108.4 "you") — the same scoping the graveyard has.
func CommanderNinjutsu(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Commander ninjutsu " + cost + " (" + cost + ", Return an unblocked attacker you control to hand: " +
			"Put this card onto the battlefield from your hand or the command zone tapped and attacking.)",
		Cost:   Plus(ManaCost(cost), ReturnAnUnblockedAttacker()),
		Zones:  []game.ZoneKind{game.ZoneHand, game.ZoneCommand},
		Effect: ninjutsuEnter,
	}
}

// ninjutsuEnter is the ability body, for both keywords. The source is
// its own target in everything but name — "this card" — so it comes
// off the stack item rather than off a target clause, which is also
// why ninjutsu cannot fizzle.
//
// Package-level rather than a closure, for the reason every delayed
// trigger body is: nothing may be captured, because undo resolves
// these against a cloned game.
func ninjutsuEnter(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	opts := game.ZoneEntryOptions{
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
	}
	// CR 608.2a: the card may have left its zone while the ability was
	// on the stack — a discard in response, a Thoughtseize. The ability
	// does as much as it can, which is nothing. Checked rather than
	// assumed, and the OWNER is checked with it: the entry door refuses
	// a card that is not in the zone it names, but a card now in
	// somebody else's hand is not this ability's card any more
	// (CR 400.7).
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Owner != item.Controller {
		return nil
	}
	// #1278, and CR 400.7 again: the card must still be the OBJECT
	// whose ability this is. A commander discarded in response goes to
	// the command zone under CR 903.9 — so it is back in a zone this
	// ability names, with the same instance ID, and a zone check alone
	// would put it onto the battlefield anyway. StackItem.SourceEpoch is
	// the announce-time identity; every zone change bumps the card's, so
	// a mismatch means the card moved and this ability lost it. It also
	// means the zone below is the zone the ability was ACTIVATED from,
	// which CR 113.6 already checked, so plain ninjutsu can never reach
	// the command-zone arm.
	if c, ok := g.LookupCardForEffect(id); !ok || c.ObjectEpoch != item.SourceEpoch {
		return nil
	}
	switch z.Kind {
	case game.ZoneHand:
		_, err := g.PutFromHandOntoBattlefieldForEffect(id, opts)
		return err
	case game.ZoneCommand:
		// CR 702.49c.
		_, err := g.PutFromCommandZoneOntoBattlefieldForEffect(id, opts)
		return err
	}
	return nil
}
