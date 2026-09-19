package game

import "github.com/google/uuid"

// adventure.go — CR 715, and ADR 0034 step 6.
//
// A card with an Adventure is ONE card with two printed faces: a
// creature (face 0) and an instant or sorcery called the Adventure
// (face 1). CR 715.3 lets the caster choose which one to cast, and
// the choice is the ordinary ADR 0034 face choice — CastableFaces
// offers both and CastSpell materialises the chosen half, exactly as
// it does for a modal DFC.
//
// What is special is the LIFECYCLE of the half that is not a
// permanent, and it is three rules:
//
//	CR 715.3d  an Adventure spell that RESOLVES is exiled instead of
//	           being put into its owner's graveyard;
//	CR 715.4   its owner may then cast it as a creature spell from
//	           exile, for as long as it stays there;
//	CR 715.3e  everything else — countered, fizzled, discarded,
//	           milled — is an ordinary card going to an ordinary
//	           graveyard, and the grant never happens.
//
// The exile is one case of the destination switch in
// routeStackCardToGraveyardLocked, beside flashback's CR 702.34a exile
// and buyback's CR 702.27b return to hand — the one place a spell
// leaving the stack picks where it goes. It is gated on that helper's
// `resolved` flag (#988), which is exactly the fact CR 715.3e turns
// on: a countered or fizzled adventure card is an ordinary card in an
// ordinary graveyard, and so is one that left the stack through the
// defensive no-StackMeta path. Before `resolved` existed this branch
// had to sit one frame up in resolveTopOfStackLocked to know the
// difference; it no longer does, and the precedence between the three
// is written down where they meet rather than spread across two files.
//
// It is BELOW #995's spellMovedItselfLocked guard and unaffected by
// it: an Adventure card exiled by CR 715.3d is not a spell that moved
// itself. The card is still on the stack when its effect finishes —
// the exile is the game's own replacement of "put it into its owner's
// graveyard", not something the card's text did — so the guard reads
// false and the frame reaches the route. A spell whose OWN text
// exiled it has already left, gets no route, and correctly gets no
// CR 715.4 grant either, because CR 715.3d never applied to it.
//
// The grant itself is ADR 0066's one model and nothing new: a
// card-scoped standing CastPermission over exile, naming the CREATURE
// face. Three of its fields carry the whole rule.
//
//	Zone: ZoneExile        where the cast comes from.
//	Duration: WhileInZoneDuration()
//	                       it needs no sweep, because the permission
//	                       names the card's CR 400.7 object epoch and
//	                       the moment the card leaves exile (cast,
//	                       Bojuka Bog'd, anything) it is a new object
//	                       the grant does not name.
//	Faces: []int{0}        "as a creature spell", CR 715.4 — the one
//	                       thing the permission could not say before
//	                       this issue, because Face was a bare int
//	                       whose zero meant "no opinion" and the
//	                       creature half IS face 0.
//
// Which is also why the restriction belongs on the PERMISSION rather
// than on the card: a Ragavan that impulse-exiles an adventure card
// opens both halves (its grant names no face), and CR 715.4's grant
// opens one. Same card, same zone, two different answers — so the
// answer cannot live on the card's own CastableFaces.

// adventureSpellFace is the face index of a card's Adventure half.
// Scryfall orders an `adventure` card's faces creature-first, which
// is also the order CR 715.2a reads them in: the card's
// characteristics are the creature's everywhere but the stack.
const adventureSpellFace = 1

// adventureCreatureFace is the face CR 715.4 opens from exile.
const adventureCreatureFace = 0

// castAsAdventure reports whether `c` is a card with an Adventure
// that is on the stack as its ADVENTURE half — the only spell CR
// 715.3d exiles.
//
// A value read off the card rather than off the stack item, because
// the face is what the rule is about: an adventure card cast as its
// creature half is an ordinary creature spell and takes the ordinary
// route to the battlefield, and one cast as its Adventure is an
// instant or sorcery whichever cost paid for it.
func castAsAdventure(c Card) bool {
	return c.Layout == LayoutAdventure && c.ActiveFace == adventureSpellFace
}

// grantAdventureCastFromExileLocked writes CR 715.4's permission onto
// the card that just landed in exile: its OWNER may cast it, as the
// creature (face 0), for as long as it remains there.
//
// The owner rather than the controller, and CR 715.4 says so in as
// many words — a Bonecrusher Giant whose Stomp was cast off somebody
// else's Ragavan grant is still its owner's creature to cast later.
//
// A card that is no longer in exile is not an error and grants
// nothing. The route's continuation runs from every terminal outcome,
// including the ones where nothing moved: an Adventure commander
// whose owner took CR 903.9's offer is in the command zone, where
// CR 715.4 has nothing to say and CR 903.4 already lets them cast it.
//
// Caller must hold g.mu (write).
func (g *Game) grantAdventureCastFromExileLocked(cardID uuid.UUID) {
	if g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID != cardID {
			continue
		}
		exiled := g.Exile.Cards[i]
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: exiled.Owner,
			Zone:   ZoneExile,
			// CR 715.4 has no end: "for as long as it remains
			// exiled". ADR 0063's vocabulary for that is
			// WhileInZone (#945), and ObjectEpoch is what enforces
			// it — see the file header.
			Duration: WhileInZoneDuration(),
			// CR 715.4: the creature, and only the creature. The
			// Adventure half was cast once and is spent.
			Faces: []int{adventureCreatureFace},
			// CR 715.4 says CAST. An adventure card's face 0 is
			// always a creature card (CR 715.2a), so nothing is
			// stranded by this; it says what the rule says.
			CastOnly: true,
			Label:    "Adventure — cast " + exiled.Name + " from exile",
		}, []Card{exiled})
		return
	}
}
