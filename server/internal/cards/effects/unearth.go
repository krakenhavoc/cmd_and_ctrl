package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// unearth.go — #1221: unearth (CR 702.82).
//
// "Unearth {1}{B}" is not an alternative way to cast the card and not
// a special action. It is an ACTIVATED ABILITY that functions only
// while the card is in its owner's GRAVEYARD (CR 702.82a):
//
//	"{1}{B}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the
//	 next end step or if it would leave the battlefield. Activate
//	 only as a sorcery."
//
// Which makes it cycling's sibling one zone over: the same CR 602
// activation, the same stack item, the same cost validation, reaching
// the graveyard through ActivatedAbilityShape.Zones (ADR 0062
// Decision 1) instead of the hand. Nothing in the engine knows the
// word "unearth".
//
// Every clause of the effect is machinery Whip of Erebos already
// built — the return, the haste grant, the end-step delayed trigger
// and the leaves-the-battlefield redirect — which is why this file is
// a constructor and a twenty-line body rather than a new primitive.
// The redirect is shared outright (exile_instead_of_leaving.go), so
// the two cards cannot drift apart on the one clause that is easy to
// get subtly wrong.

// Unearth is "Unearth <cost>" — CR 702.82a. `cost` is the printed
// mana cost of the ability, "{1}{B}" or "{3}{R}{R}"; the graveyard
// zone, the sorcery-speed restriction and the whole of the return
// come with the keyword and are not the card file's to spell out.
//
// Give the ability to the card the way its oracle text reads:
//
//	Activated: []ActivatedAbility{Unearth("{1}{B}")},
func Unearth(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Unearth " + cost + " (" + cost + ": Return this card from your graveyard to the battlefield. " +
			"It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. " +
			"Activate only as a sorcery.)",
		Cost:         ManaCost(cost),
		Zones:        []game.ZoneKind{game.ZoneGraveyard},
		SorcerySpeed: true,
		Effect:       unearthReturn,
	}
}

// unearthReturn is the ability body. The source is its own target in
// everything but name — "this card" — so it comes off the stack item
// rather than off a target clause, which is also why unearth cannot
// fizzle.
//
// Package-level rather than a closure for the reason every delayed
// trigger body is: nothing may be captured, because undo resolves
// these against a cloned game.
func unearthReturn(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	ctx := NewContext(g, item)
	// CR 608.2a: the card may have left the graveyard while the
	// ability was on the stack (a second player's graveyard hate, a
	// Bojuka Bog). The ability does as much as it can, which is
	// nothing. Checked rather than assumed: ReturnFromGraveyard on a
	// card in exile would be a reanimation the card never printed.
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if err := (ReturnFromGraveyard{
		Target: id,
		Dest:   game.ZoneBattlefield,
		// CR 108.4 made the activator the card's owner, so "under
		// your control" and "under its owner's control" are the same
		// player here — named anyway, because the rule the card
		// prints is the first one.
		Controller: item.Controller,
	}).Apply(ctx); err != nil {
		return err
	}
	// An entry replacement may have sent it somewhere else
	// (a Leyline of the Void's exile, a Containment Priest's), and
	// the three clauses below are all about a permanent that entered.
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Unearth — the returned creature has haste",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (ScheduleDelayedTrigger{
		Label: "Unearth — exile the returned creature",
		Cards: []uuid.UUID{id},
		Body:  exileListedCardsBody,
	}).Apply(ctx); err != nil {
		return err
	}
	ExileInsteadOfLeavingBattlefield(g, id, item.Controller, "Unearth: if it would leave the battlefield, exile it instead")
	return nil
}
