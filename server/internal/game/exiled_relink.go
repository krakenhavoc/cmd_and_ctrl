package game

import "github.com/google/uuid"

// exiled_relink.go — CR 607.2a for a card whose link has to name an object
// that did not exist when the card was exiled (Timothar, Baron of Bats: "you
// may pay {1} and exile it. If you do, create a 1/1 black Bat creature
// token. It gains 'When this token deals combat damage to a player, sacrifice
// it and return the exiled card to the battlefield tapped.'"). The card is
// exiled first and the Bat made after it, so the exiled card is linked to the
// Bat once the Bat exists.

// RelinkExiledWithForEffect makes the card in exile `cardID` "exiled with"
// the object `to`, replacing any link it had, and reports whether the card
// is in exile. It reads and writes the exile zone only and changes no
// snapshot shape: Card.ExiledWith is already persisted.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) RelinkExiledWithForEffect(cardID uuid.UUID, to PermissionCardRef) bool {
	if g.Exile == nil {
		return false
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == cardID {
			g.Exile.Cards[i].ExiledWith = to
			g.Exile.Cards[i].ExiledWithAbility = AbilityRef{}
			return true
		}
	}
	return false
}
