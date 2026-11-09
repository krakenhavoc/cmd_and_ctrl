package game

import "github.com/google/uuid"

// cloak.go — CR 701.58, "cloak" (#4452: Etrata, Deadly Fugitive; Vannifar,
// Evolved Enigma; Hide in Plain Sight).
//
// To cloak a card, its controller puts it onto the battlefield face down as
// a 2/2 creature with ward {2} (CR 701.58a): manifest (CR 701.40a) with the
// ward, which is the FaceDownCloaked state's, so the engine needs no door of
// its own. The ward, the 2/2 projection, the catalog suppression, the viewers
// rule and the turn-face-up offer by the card's mana cost (CR 701.58b, a
// creature card only; a morph or disguise cost besides, CR 701.58c) are all
// face_down.go's. What cloak adds is who it is done to, and three doors:
//
//   - the top card of a library, any player's, under the cloaker's control:
//     CloakTopOfLibraryForEffect ("cloak the top card of that player's
//     library", Etrata, Deadly Fugitive);
//   - cards chosen from a hand: PutCardsFromHandOntoBattlefieldKeyedForEffect
//     with HandEntryOptions{FaceDown: FaceDownCloaked} ("cloak a card from
//     your hand", Vannifar);
//   - cards chosen from a library: PutCardsFromLibraryOntoBattlefieldKeyedForEffect
//     with LibraryEntryOptions{Controller: you, FaceDown: FaceDownCloaked}
//     ("look at the top five cards, cloak two of them", Hide in Plain Sight).
//
// The card is not revealed on its way in, and the entering permanent runs no
// enters-the-battlefield trigger of its own (CR 708.2a), exactly as for
// manifest.

// CloakTopOfLibraryForEffect cloaks the top card of libraryOwner's library
// under controller's control and returns the permanent, or uuid.Nil with a
// nil error when the library is empty or a replacement canceled, redirected
// or paused the entry (a keyed caller waits for it through the entry's own
// tail instead). The two players are one for "cloak the top card of your
// library"; they differ for Etrata's "cloak the top card of that player's
// library".
//
// Caller must hold g.mu.
func (g *Game) CloakTopOfLibraryForEffect(controller, libraryOwner uuid.UUID) (uuid.UUID, error) {
	if g.playerByIDLocked(controller) == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	p := g.playerByIDLocked(libraryOwner)
	if p == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	if p.Library.Size() == 0 {
		// An empty library cloaks nothing and is not a draw.
		return uuid.Nil, nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{top}, ZoneLibrary, ZoneEntryOptions{
		Controller: controller,
		FaceDown:   FaceDownCloaked,
	})
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}
