package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reassembling Skeleton — Creature — Skeleton Warrior {B}, 1/1
// (EDHREC rank 761):
//
//	"{1}{B}: Return this card from your graveyard to the battlefield
//	 tapped."
//
// A CR 602 activation from the graveyard (CR 113.6), reaching it
// through ActivatedAbilityShape.Zones exactly as cycling reaches a
// hand and Unearth reaches a graveyard one file over (#1221) — the
// same activation path, the same cost validation, the same stack
// item. No sorcery-speed restriction is printed, unlike Unearth, so
// none is declared.
//
// The one clause the ability has — "enters tapped" — is
// ReturnFromGraveyard.Tapped (#1284), which rides the CR 614 entry
// exactly as SearchLibrary.TappedOnEntry does one primitive over: the
// permanent enters tapped, full stop, rather than entering and then
// being tapped a beat later (the two are observably different events;
// see the primitive's own doc comment).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9dbc3530-b278-4c8d-b2cc-a09dfac9d5e5",
		Name:            "Reassembling Skeleton",
		Completeness:    CompletenessFull,
		PrintedKeywords: nil,
		Activated: []ActivatedAbility{{
			Label: "{1}{B}: Return this card from your graveyard to the battlefield tapped.",
			Cost:  ManaCost("{1}{B}"),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardTapped(g, item)
			},
		}},
	})
}

// returnThisFromGraveyardTapped is the shared body for "<cost>:
// Return this card from your graveyard to the battlefield tapped." —
// Reassembling Skeleton (above) and Drownyard Temple (#1284). The
// whole printed ability IS the entry: no haste grant, no
// leaves-the-battlefield redirect, none of Unearth's other clauses.
//
// Package-level rather than a closure captured at each call site, for
// the reason every delayed trigger body and unearthReturn already
// are: nothing may be captured, because undo resolves an ability's
// Effect against a cloned game.
func returnThisFromGraveyardTapped(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	// CR 602.5 / 608.2a: the card may have left the graveyard between
	// activation and resolution (a second player's graveyard hate, a
	// Bojuka Bog naming this graveyard). The ability does as much as
	// it can, which is nothing — reanimating a card now sitting in
	// exile would be a return the card never printed.
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return (ReturnFromGraveyard{
		Target: id,
		Dest:   game.ZoneBattlefield,
		// CR 602.2 made the activator the ability's controller, which
		// for a card activating an ability from ITS OWN graveyard is
		// also its owner — named anyway, because the rule the card
		// prints ("the battlefield", no "under your control" clause)
		// is CR 400.3's default: the owner's control.
		Controller: item.Controller,
		Tapped:     true,
	}).Apply(NewContext(g, item))
}
