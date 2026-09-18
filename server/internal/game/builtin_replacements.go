package game

import "github.com/google/uuid"

// builtin_replacements.go holds the engine-owned replacement
// effects — the ones that live in the rules themselves, not on
// any particular card. Registered on every game at NewGame.
//
// Two today: commanderZoneReplacement (CR 903.9), refactored from
// S13.1's inline applyCommanderZoneReplacementLocked, and
// regenerationShieldReplacement (CR 701.19). Future built-ins (e.g.
// the "if a card would be exiled by an effect that replaces it with
// something else" meta-replacement) would live here too.
//
// What makes a rule a BUILT-IN rather than a card's declared
// effect: it is printed in the Comprehensive Rules rather than on any
// object, and it has no source card for the pipeline to hang it on.
// CR 903.9 belongs to the format; a regeneration shield belongs to the
// permanent that was given one and outlives the ability that made it
// — the Asceticism that regenerated your creature may be gone by the
// time the shield is spent, and the shield does not care.
//
// A built-in is registered ONCE per game, so two shields on one
// permanent are still ONE applicable replacement in the CR 616 gather:
// there is no ordering prompt to ask about, and the second shield
// waits for the next destruction. See replacementIdentity — a built-in
// deliberately has none, because two registrations of one built-in
// would not be two printings of one effect.
//
// Added in S17 sub-PR 2.

// commanderZoneReplacement implements CR 903.9: "if a commander
// would be put into a library, hand, graveyard, or exile from
// anywhere, its owner may put it into the command zone instead."
//
// CR 614.10 "may" replacement — Optional=true triggers the apply-
// loop's yes/no prompt path (PendingChoiceOptionalReplacement) so
// the owner decides each time. Sub-PR 6 widened AppliesTo (dropped
// the asCommanderMove gate) so this fires for every move that puts a
// commander into graveyard/exile/hand/library, whatever sent it.
//
// A destination-only AppliesTo is necessary but not sufficient: a
// replacement effect only ever runs for a mover that pushes a
// RepEventMove through applyReplacementsLocked. Until #529 only two
// callers did, both battlefield → graveyard, so this widening was
// unreachable from counter, fizzle, exile, bounce, tuck and mill —
// which is most of how a commander leaves in a real game. The
// window now lives in the shared exit primitive
// (routeCardToZoneLocked, zone_route.go) rather than in the movers,
// so "from anywhere" holds for every route that goes through it.
//
// Controlled by the commander's owner (drives both the CR 614.10
// yes/no prompt and the CR 616 multi-replacement order prompt if
// other commander-zone-touching replacements ever join).
var commanderZoneReplacement = ReplacementEffect{
	// Two kinds, because #650 split the discard off: "from anywhere"
	// has to include a discarded commander, and a discard no longer
	// arrives as a zone move.
	Watches:        []EventKind{EventZoneMove, EventDiscardCard},
	Optional:       true,
	PromptQuestion: "Send commander to command zone instead?",
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if !isExitMove(ev.Kind) {
			return false
		}
		switch ev.NewZone {
		case ZoneGraveyard, ZoneExile, ZoneHand, ZoneLibrary:
			// Eligible destination.
		default:
			return false
		}
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		return card.IsCommander
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		ev.NewZone = ZoneCommand
		ev.NewZoneOwner = card.Owner
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return uuid.Nil
		}
		return card.Owner
	},
	Label: "Commander zone replacement",
}

// regenerationShieldReplacement implements CR 701.19a: "The next time
// that permanent would be destroyed this turn, instead its controller
// taps it, removes all damage marked on it, and removes it from
// combat."
//
// The rule, the shield representation and everything the shield DOES
// live in regeneration.go; this is the pipeline registration, written
// the way commanderZoneReplacement above is so the two engine-owned
// replacements read as siblings.
//
// NOT PureCancel, although its Replace does call ev.Cancel(). That
// flag is a declaration that Replace "rewrites no other field on the
// event and changes nothing else in the game", and this one changes
// four things about the permanent. Leaving it false costs a CR 616
// ordering prompt in the one window where a second replacement also
// applies — a COMMANDER with a shield on it, where CR 903.9 is also
// offering — and that prompt is the rules-correct outcome: CR 616.1
// gives the affected permanent's controller the choice, and the two
// orders are genuinely different questions to be asked. (They reach
// the same board either way, as it happens: regeneration first cancels
// the move outright, and CR 903.9 first rewrites the destination of a
// move that regeneration then cancels anyway under the CR 616.1f
// re-check.) Two SHIELDS never prompt, because they are one
// registration of one built-in — see the file header.
//
// Controlled by the permanent's controller: a regeneration shield
// is created for a permanent and CR 701.19a hands the actions to "its
// controller", so that is who CR 616 asks about ordering.
var regenerationShieldReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.regenerationShieldAppliesLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		g.applyRegenerationShieldLocked(ev)
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		return g.controllerOfBattlefieldCardLocked(ev.CardID)
	},
	Label: "Regeneration shield",
}
