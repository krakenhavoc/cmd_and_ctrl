package game

import "github.com/google/uuid"

// manifest_library.go — manifest from a library that is not the
// manifesting player's own (#5429: Thieving Amalgam, Orochi Soul-Reaver).
//
// "You manifest the top card of that player's library" has two players:
// the one who manifests, and so controls the face-down permanent
// (CR 110.2a: "If an effect instructs a player to put an object onto the
// battlefield, that object enters the battlefield under that player's
// control unless the effect states otherwise"), and the one whose library
// it comes from, who stays its owner (CR 108.3). ManifestForEffect
// (battlefield_put.go) is the door where the two are one player; cloak has
// had the split since Etrata (CloakTopOfLibraryForEffect, cloak.go), and
// this is the same split for CR 701.40a.
//
// Everything a manifested permanent is comes from the FaceDownManifested
// entry unchanged: the 2/2 projection and the suppressed catalog
// (CR 708.2a), the controller as the only player who may look at it
// (CR 708.5: "You can't look at ... face-down spells or permanents
// controlled by another player", so the library's owner may NOT), and the
// turn-face-up offer for the card's mana cost, the controller's to take
// (CR 701.40b: "you may turn a manifested permanent you control face up").

// ManifestTopOfLibraryForEffect manifests the top card of libraryOwner's
// library under controller's control (CR 701.40a) and returns the
// permanent, or uuid.Nil with a nil error when the library is empty or a
// replacement canceled, redirected or paused the entry. The card is not
// revealed on the way, and it need not be a permanent card.
//
// Caller must hold g.mu.
func (g *Game) ManifestTopOfLibraryForEffect(controller, libraryOwner uuid.UUID) (uuid.UUID, error) {
	if g.playerByIDLocked(controller) == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	p := g.playerByIDLocked(libraryOwner)
	if p == nil {
		return uuid.Nil, ErrPlayerNotFound
	}
	if p.Library.Size() == 0 {
		// An empty library manifests nothing and is not a draw.
		return uuid.Nil, nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	entered, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{top}, ZoneLibrary, ZoneEntryOptions{
		Controller: controller,
		FaceDown:   FaceDownManifested,
	})
	if err != nil || len(entered) == 0 {
		return uuid.Nil, err
	}
	return entered[0], nil
}
