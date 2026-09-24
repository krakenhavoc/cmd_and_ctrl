package game

import "github.com/google/uuid"

// regeneration.go teaches the engine CR 701.19 — the keyword action
// that, until #667, the catalog could print and nothing could read.
//
// # What regeneration is, in the rules' own words
//
// CR 701.19a: "If the effect of a resolving spell or ability
// regenerates a permanent, it creates a shield for that permanent. …
// The next time that permanent would be destroyed this turn, instead
// its controller taps it, removes all damage marked on it, and
// removes it from combat."
//
// Four things are worth pulling out of that sentence, because each
// one decides a line of code below:
//
//   - It is a REPLACEMENT EFFECT on the destruction, not a trigger
//     and not a state-based action. So it lives where every other
//     replacement lives — the CR 614 pipeline — and it lives there as
//     an ENGINE built-in (builtin_replacements.go), not as a card's
//     declared effect, because the shield outlives the ability that
//     made it and belongs to the permanent rather than to a source.
//   - It is a SHIELD with a charge. "The next time" is one
//     destruction; regenerating twice is two shields and survives two
//     destructions. So the state is a COUNT on the permanent —
//     Card.RegenerationShields — decremented when it applies.
//   - It lasts UNTIL END OF TURN. Cleared by the cleanup sweep
//     (CR 514.2, rotation.go) beside the marked damage it is
//     bookkeeping-shaped like, and by the battlefield exit, because
//     CR 400.7 makes what lands in the next zone a new object.
//   - What it DOES is three things and only three: tap, remove all
//     damage, remove from combat. The permanent is not destroyed, so
//     it does not leave the battlefield, nothing "dies", and no
//     dies-trigger fires.
//
// # What regeneration does NOT stop
//
// This list is the reason the rider is declared on the event rather
// than inferred from "a permanent is on its way to a graveyard":
//
//	CR 701.21a  SACRIFICE. Never destruction. sacrifice.go routes
//	            through the same battlefield exit this does, and the
//	            ONLY thing that tells the two apart is the
//	            Destruction flag the destroy route sets.
//	CR 704.5f   toughness 0 or less — the creature is PUT INTO a
//	            graveyard, not destroyed.
//	CR 704.5i   a planeswalker at 0 loyalty — likewise.
//	CR 704.5v/w   a battle at 0 defense — sacrificed.
//	CR 704.5j   the legend rule — put into a graveyard.
//	CR 704.5m   an illegally attached Aura — put into a graveyard.
//	            Exile, bounce, tuck and mill, obviously.
//
// The SBA sweep is the awkward one, because ONE pass collects
// permanents doomed by all of CR 704.5f/g/h/i/v and destroys them as
// one simultaneous event (CR 704.3). Two of those arms are
// destruction and three are not, so the sweep carries the flag per
// permanent — see doomedPermanent in simultaneous.go.
//
// # "Can't be regenerated"
//
// CR 701.19c: an effect that says a permanent can't be regenerated
// ignores regeneration shields. CR 701.19c adds the part that is easy
// to get wrong: the shield is NOT used up. Damnation destroys a
// creature with a shield on it; if somehow it survived, the shield
// would still be there for the next destruction. So the rider gates
// AppliesTo rather than being consumed inside Replace.
//
// It rides as one bool from the destroying effect down to the event:
//
//	DestroyOptions{CantBeRegenerated: true}
//	  → destroyRouteWith(opts)        (the zoneRoute template)
//	  → ReplacementEvent.CantBeRegenerated
//	  → the built-in's AppliesTo says no.
//
// # Where the pieces are
//
//	Card.RegenerationShields        card.go — the shield count
//	RegenerateForEffect             here — "regenerate [permanent]"
//	regenerationShieldReplacement   builtin_replacements.go — the rule
//	applyRegenerationShieldLocked   here — what the shield DOES
//	clearRegenerationShieldsLocked  here, from rotation.go's cleanup
//	DestroyOptions                  here — the can't-be-regenerated rider
//
// Added in #667.

// DestroyOptions is the rider a destruction may print beyond "destroy
// it". One field today: CR 701.19c's "it can't be regenerated", which
// every wipe and kill spell that prints the clause passes.
//
// A struct rather than a bare bool because the destroy entry points
// are the catalog's most-used verbs and a second positional boolean
// at every call site would be unreadable; and because totem armor
// (CR 702.111) and "if it would die, exile it instead" are the same
// shape of rider when they land.
type DestroyOptions struct {
	// CantBeRegenerated makes this destruction ignore regeneration
	// shields (CR 701.19c). The shields are not spent — CR 701.19c
	// leaves them on the permanent for a later destruction that does
	// not say this.
	CantBeRegenerated bool
}

// firstDestroyOptions reads the optional rider off a variadic
// ...DestroyOptions parameter.
//
// The destroy verbs take the rider variadically because ~120 call
// sites in the engine, the catalog and their tests say "destroy this"
// with nothing to add, and making every one of them write an empty
// struct would be a hundred lines of noise around one real change.
// At most one is meaningful; a second is ignored rather than merged,
// because "destroy it twice, and the second time it can't be
// regenerated" is not a thing any card says.
func firstDestroyOptions(opts []DestroyOptions) DestroyOptions {
	if len(opts) == 0 {
		return DestroyOptions{}
	}
	return opts[0]
}

// RegenerateForEffect creates one regeneration shield for a
// permanent (CR 701.19a) — the whole of what "Regenerate target
// creature" does when it resolves. The shield sits on the permanent
// until it replaces a destruction or until the cleanup step, and a
// second call stacks a second shield.
//
// Returns ErrCardNotFound for a card that is not on the battlefield:
// CR 701.19b says regenerating a permanent that is not there does
// nothing, and a targeted regeneration whose target has already left
// is the ordinary way to reach that.
//
// It deliberately does NOT check whether the permanent could ever be
// destroyed. Regenerating an indestructible creature, or a land, is
// legal and wasteful, exactly as it is in paper.
//
// Caller must hold g.mu (a *ForEffect surface — it runs inside a
// resolution that already owns the write lock).
func (g *Game) RegenerateForEffect(cardID uuid.UUID) error {
	idx := findCardOnBattlefield(g, cardID)
	if idx < 0 {
		return ErrCardNotFound
	}
	g.Battlefield.Cards[idx].RegenerationShields++
	return nil
}

// RegenerationShieldsOn reports how many regeneration shields a
// permanent is carrying. Zero for a card that is not on the
// battlefield.
//
// Exported for the wire view and for catalog tests; the rule itself
// reads the field directly.
//
// Caller must hold g.mu (read or write).
func (g *Game) RegenerationShieldsOn(cardID uuid.UUID) int {
	if idx := findCardOnBattlefield(g, cardID); idx >= 0 {
		return g.Battlefield.Cards[idx].RegenerationShields
	}
	return 0
}

// regenerationShieldAppliesLocked is the built-in's AppliesTo: does
// this permanent have a shield, and is this event a destruction the
// shield is allowed to replace?
//
// Three conditions, one per clause of the rule:
//
//	ev.Destruction        CR 701.19a — a shield replaces a
//	                      DESTRUCTION. A sacrifice, a zero-toughness
//	                      SBA, an exile or a bounce takes the same
//	                      battlefield exit and is not one.
//	!ev.CantBeRegenerated CR 701.19c — the destroying effect said no.
//	                      Gated here rather than inside Replace so
//	                      the shield is not spent (CR 701.19c).
//	shields > 0           there is a shield to spend.
//
// Caller must hold g.mu.
func (g *Game) regenerationShieldAppliesLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventMove || !ev.Destruction || ev.CantBeRegenerated {
		return false
	}
	c := findBattlefieldCard(g, ev.CardID)
	return c != nil && c.RegenerationShields > 0
}

// applyRegenerationShieldLocked is the built-in's Replace: spend one
// shield and perform the regeneration (CR 701.19a).
//
// The destruction is CANCELLED rather than redirected. "Instead" with
// a null zone change is exactly what ev.Cancel() means everywhere
// else in this pipeline, and it is what makes the rest fall out
// right: the permanent never leaves, so MoveCard's battlefield-exit
// cleanup never runs, no EventLTB fires, nothing dies, and
// destroyedThisWayLocked reads the live board and counts no
// destruction (#815).
//
// Then the three things the rule names, in its order:
//
//   - TAP IT. Not "tap it if it is untapped" — a tapped permanent
//     simply stays tapped, which is what assigning true does.
//   - REMOVE ALL DAMAGE. Through the same clearBattlefieldDamage the
//     cleanup step and the battlefield exit use, so "what clearing
//     means" — the damage AND the CR 702.2c deathtouch mark — is
//     written down once (#708 / #816). This is also what stops the
//     lethal-damage state-based action (CR 704.5g) from destroying
//     the creature again on the very next pass: the damage that
//     doomed it is gone, so the SBA finds nothing. Without it a
//     regenerated creature would spend its shields in a loop and die
//     anyway.
//   - REMOVE IT FROM COMBAT. Through #921's removeFromCombatLocked,
//     which clears the attack and block ANNOUNCEMENTS as well as the
//     declarations (CR 506.4, #871).
//
// The shield is decremented FIRST, so a Replace that somehow re-enters
// the pipeline cannot spend the same shield twice.
//
// Caller must hold g.mu.
func (g *Game) applyRegenerationShieldLocked(ev *ReplacementEvent) {
	if ev == nil {
		return
	}
	idx := findCardOnBattlefield(g, ev.CardID)
	if idx < 0 {
		return
	}
	c := &g.Battlefield.Cards[idx]
	if c.RegenerationShields <= 0 {
		return
	}
	c.RegenerationShields--
	ev.Cancel()
	c.Tapped = true
	clearBattlefieldDamage(c)
	g.removeFromCombatLocked(c)
	g.EmitEvent(Event{
		Kind:   EventRegenerated,
		Actor:  c.Controller,
		Source: c.InstanceID,
		CardID: c.InstanceID,
	})
}

// clearRegenerationShieldsLocked drops every unused shield on the
// board. CR 701.19a scopes a shield to the turn it was created in, so
// the cleanup step's sweep is where it ends (rotation.go,
// sweepTurnEndLocked) — the same pass that removes marked damage, and
// for the same reason: both are per-turn state on one permanent.
//
// Idempotent, like everything else in that sweep.
//
// Caller must hold g.mu.
func (g *Game) clearRegenerationShieldsLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].RegenerationShields = 0
	}
}
