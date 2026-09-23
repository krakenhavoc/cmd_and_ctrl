package game

import "github.com/google/uuid"

// builtin_replacements.go holds the engine-owned replacement
// effects — the ones that live in the rules themselves, not on
// any particular card. Registered on every game at NewGame.
//
// Three today: commanderZoneReplacement (CR 903.9), refactored from
// S13.1's inline applyCommanderZoneReplacementLocked,
// regenerationShieldReplacement (CR 701.19), and
// protectionPreventsDamageReplacement (CR 702.16e). Future built-ins
// (e.g. the "if a card would be exiled by an effect that replaces it
// with something else" meta-replacement) would live here too.
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

// protectionPreventsDamageReplacement implements CR 702.16e: "any
// damage that would be dealt by sources that have the stated quality
// to a permanent or player with protection from that quality is
// prevented."
//
// The third engine-owned replacement, and a built-in for the same
// reason the other two are: it is printed in the Comprehensive Rules
// rather than on any object. The protection is an ability of the
// permanent being damaged, but the PREVENTION is a rule, and there is
// no card for the pipeline to hang it on.
//
// PREEMPTIVE. It applies before every other applicable replacement
// and asks no CR 616 ordering question. That is what stops a charged
// prevention shield (effects.PreventNextDamage, "prevent the next 4
// damage") from spending a charge absorbing damage that was never
// going to be dealt — #420 — and it is a declared simplification of
// CR 616.1, which would hand the ordering choice to the affected
// permanent's controller. ADR 0072 §4 carries the argument and names
// the cost (Phytohydra, which would rather its own replacement
// applied).
//
// THE SOURCE IS READ FROM LAST-KNOWN INFORMATION, never from a
// battlefield lookup. ev.SourceLKI is the source's characteristics as
// they were when the damage event was created (CR 608.2h), which is
// the only thing that answers for a SPELL — never on the battlefield
// at all — and for an attacker already swept into a graveyard by the
// time a CR 616 pause resumes. A nil LKI is an unknown source and
// prevents nothing, which errs weaker.
//
// PLAYERS ARE COVERED since #1197. ev.DamageTarget is a card ID or a
// player ID, and the branch that used to say "a player has no ability
// slice to read" now asks the player-side reader instead
// (player_statics.go). CR 702.16e names a permanent and a player in
// one sentence, so this is one rule with two stores rather than two
// rules. ADR 0072's 2026-09-22 amendment.
//
// CR 615.12 ("this damage can't be prevented") is not modelled
// anywhere in this engine, so it does not stop this either — the
// pre-existing gap Banefire's caveat already names.
//
// No Controller: CR 616.1 would ask the affected permanent's
// controller, and a preemptive effect is never ordered against
// anything, so there is nobody to ask.
var protectionPreventsDamageReplacement = ReplacementEffect{
	Watches:    []EventKind{EventDealDamage},
	Preemptive: true,
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.protectionPreventsDamageLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		// CR 615.1: prevented damage is not dealt at all — no
		// EventDealDamage, no lifelink, no deathtouch, no "whenever ~
		// is dealt damage" trigger. Cancel is exactly that.
		ev.Cancel()
		return nil
	},
	Label: "Protection",
}

// protectionPreventsDamageLocked is the built-in's AppliesTo: is the
// damage target a battlefield permanent with protection from the
// source this event's last-known information describes?
//
// Caller must hold g.mu.
func (g *Game) protectionPreventsDamageLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return false
	}
	// A source-less event (a sandbox mark with no source, an effect
	// that names none) prevents nothing: CR 702.16e is about a source
	// WITH the quality, and an unknown one is not known to have it.
	if ev.SourceLKI == nil {
		return false
	}
	idx := findCardOnBattlefield(g, ev.DamageTarget)
	if idx < 0 {
		// Not a permanent: a PLAYER, or a permanent that has already
		// left. CR 702.16e names both in one sentence — "damage that
		// would be dealt … to a permanent or player with protection
		// from that quality is prevented" — so the player branch is
		// this same rule and not a second one (#1197, ADR 0072's
		// 2026-09-22 amendment).
		//
		// ONE BRANCH COVERS BOTH KINDS OF DAMAGE. Every damage entry
		// point in the engine — combat, trample overflow, a spell, an
		// ability — routes through damageThroughReplacementsLocked,
		// which is where SourceLKI is stamped. A player with
		// protection from everything is therefore shielded from a
		// Lightning Bolt and from an attacking 8/8 by the same four
		// lines.
		//
		// A player who is not a seat (a stale ref, a permanent that
		// left) finds no player and prevents nothing, which errs
		// weaker.
		return g.playerProtectionPreventsDamageLocked(ev)
	}
	c := &g.Battlefield.Cards[idx]
	return HasProtection(c) && ProtectedFrom(c, ev.SourceLKI)
}

// playerProtectionPreventsDamageLocked is the player half of
// CR 702.16e: is the damage target a seat with protection from the
// source this event's last-known information describes?
//
// Split out rather than inlined because the two halves read different
// stores — the battlefield for a permanent's abilities, the derived +
// granted reader for a player's — and the branch above is already the
// one place that tells them apart.
//
// Caller must hold g.mu.
func (g *Game) playerProtectionPreventsDamageLocked(ev *ReplacementEvent) bool {
	p := g.playerByIDLocked(ev.DamageTarget)
	if p == nil {
		return false
	}
	_, prevented := g.PlayerProtectedFromLocked(p, ev.SourceLKI)
	return prevented
}
