package game

import "github.com/google/uuid"

// hideaway.go — CR 702.75, hideaway (ADR 0091, #1331).
//
//	"Hideaway N" means "When this permanent enters, look at the top N
//	 cards of your library. Exile one of them face down and put the
//	 rest on the bottom of your library in a random order. The exiled
//	 card gains 'The player who controls the permanent that exiled this
//	 card may look at this card in the exile zone.'"   — CR 702.75a
//
// And every hideaway card then prints a linked ability (CR 607.2a):
// "… you may play the exiled card without paying its mana cost".
//
// Three things the engine had no shape for, each small:
//
//  1. A FACE-DOWN EXILE WHOSE VIEWER IS ANOTHER OBJECT'S CONTROLLER —
//     FaceDownHidden. ADR 0069's viewers table had "nobody" (a plain
//     face-down exile), "the owner" (foretell) and "the controller"
//     (a CR 708.2 permanent, whose controller is its own). Hideaway's
//     viewer is the controller of the PERMANENT THAT EXILED IT, so the
//     answer is read through Card.HiddenBy (faceDownViewersLocked) and
//     kept current as that permanent changes hands
//     (hideawayKnowersSweepLocked). CR 406.3 keeps anyone who has
//     already looked.
//
//  2. THE LINK — Card.HiddenBy, the permanent OBJECT {instance, epoch},
//     stamped as the card lands and cleared by MoveCard on any move.
//     HiddenCardsForEffect is CR 607.2a's "the exiled card": the cards
//     still in exile that THIS object's hideaway put there.
//
//  3. THE FREE PLAY is not engine work at all: it is an ADR 0066 grant
//     over the linked card, "{0}" and not cast-only (hideaway says
//     PLAY — a land may be hidden and played), written by the card's
//     own linked ability (effects.PlayHiddenCard). The engine already
//     casts a face-down exiled card (foretell, CR 406.3a turns it face
//     up as it is cast); a land played from exile comes through the
//     same land branch.
//
// The look-and-choose itself is effects.Hideaway, which composes the
// primitives that already exist: LookAtTopOfLibraryForEffect, a
// choose_cards prompt, ExileHiddenForEffect below, and
// PutOnBottomInRandomOrderForEffect.

// ObjectRefOf is the {instance, CR 400.7 epoch} pair that names one
// OBJECT rather than one card — the shape a hideaway link, a prepare
// copy's permanent and a cast permission's card all use.
func ObjectRefOf(c Card) PermissionCardRef {
	return PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
}

// hiddenSourceLocked is the permanent a hideaway link names, while it
// is still that object on the battlefield; nil otherwise.
func (g *Game) hiddenSourceLocked(ref PermissionCardRef) *Card {
	if ref.ID == uuid.Nil {
		return nil
	}
	src := findBattlefieldCard(g, ref.ID)
	if src == nil || src.ObjectEpoch != ref.Epoch {
		return nil
	}
	return src
}

// hiddenByControllerLocked answers FaceDownHidden's viewer rule: the
// controller of the permanent that exiled the card (CR 702.75a), or
// uuid.Nil when that permanent is gone.
func (g *Game) hiddenByControllerLocked(c Card) uuid.UUID {
	if src := g.hiddenSourceLocked(c.HiddenBy); src != nil {
		return src.Controller
	}
	return uuid.Nil
}

// ExileHiddenForEffect exiles one card face down under hideaway
// (CR 702.75a): FaceDownHidden, linked to `source` — the permanent
// object whose hideaway is resolving — so its controller may look at
// it and its linked ability can find it. Through the shared exit
// primitive, so a commander hidden off the top of its owner's library
// is still offered the command zone (CR 903.9); a card that goes there
// instead was never hidden and carries no link.
//
// Answers whether the move paused on such a prompt.
//
// Caller must hold g.mu.
func (g *Game) ExileHiddenForEffect(source PermissionCardRef, actor, cardID uuid.UUID) (bool, error) {
	return g.routeCardToZoneLocked(zoneRoute{
		CardID:   cardID,
		Dst:      ZoneExile,
		Actor:    actor,
		Source:   source.ID,
		FaceDown: FaceDownHidden,
		HiddenBy: source,
	})
}

// HiddenCardsForEffect is CR 607.2a's "the exiled card" for a hideaway
// permanent: the cards still in exile that `source` — that exact
// object — exiled with its hideaway, oldest first. Empty once the card
// has been played or has left exile any other way, and empty for a
// hideaway permanent that was bounced and replayed (a new object hid
// nothing). Evercoat Ursine's two hideaways make two.
//
// `source` is an object reference rather than an instance ID so a
// linked ability resolving after its permanent has left the
// battlefield still finds what that permanent hid (CR 607.2a reads the
// exile, not the permanent): the caller captures the reference as the
// ability goes on the stack.
//
// Caller must hold g.mu.
func (g *Game) HiddenCardsForEffect(source PermissionCardRef) []uuid.UUID {
	if g.Exile == nil || source.ID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.HiddenBy == source {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// hideawayKnowersSweepLocked keeps FaceDownHidden's viewer current: the
// controller of the permanent that hid each card is a knower of it.
// A hideaway land that changes hands hands the look over with it; the
// old controller is not removed, because CR 406.3 lets a player who
// has looked "continue to look at that card until it leaves the exile
// zone … even if the instruction allowing the player to do so no
// longer applies".
//
// Run with the state checks, where every other "the board changed, is
// this still true" rule runs. It changes no game object's
// characteristics and fires nothing, so it never extends the SBA loop.
//
// Caller must hold g.mu in write mode.
func (g *Game) hideawayKnowersSweepLocked() {
	if g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.FaceDownKind != FaceDownHidden || c.HiddenBy.ID == uuid.Nil {
			continue
		}
		if viewer := g.hiddenByControllerLocked(*c); viewer != uuid.Nil && g.playerByIDLocked(viewer) != nil {
			c.AddKnower(viewer)
		}
	}
}
