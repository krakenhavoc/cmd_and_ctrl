package effects

import (
	"fmt"

	"github.com/google/uuid"
)

// cloak_action.go — cloak (CR 701.58), the card side of game/cloak.go.
//
// CloakTopOfLibrary is "cloak the top card of your library": the card
// enters face down as a 2/2 creature with ward {2} (manifest's sibling,
// manifest.go). Its New field is the permanent that entered (uuid.Nil on
// an empty library), for "then attach this Equipment to it" (Cryptic
// Coat) and "when you do" shapes.
//
// Cards that cloak cards from elsewhere (Expose the Culprit) call
// game.CloakCardsKeyedForEffect directly.

// CloakTopOfLibrary cloaks the top card of a library: Player's, under
// Player's control (the resolving item's controller when zero). Controller
// and LibraryOwner split the two for "cloak the top card of that player's
// library" (Etrata, Deadly Fugitive): Controller is who puts it onto the
// battlefield and controls it (zero: Player as above), LibraryOwner whose
// library it comes from (zero: the controller).
type CloakTopOfLibrary struct {
	Player       uuid.UUID
	Controller   uuid.UUID
	LibraryOwner uuid.UUID
}

// Do cloaks the card and returns the permanent that entered.
func (c CloakTopOfLibrary) Do(ctx *Context) (uuid.UUID, error) {
	controller := c.Controller
	if controller == uuid.Nil {
		controller = c.Player
	}
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	owner := c.LibraryOwner
	if owner == uuid.Nil {
		owner = controller
	}
	id, err := ctx.Game.CloakTopOfLibraryForEffect(controller, owner)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cloak: %w", err)
	}
	return id, nil
}

// Apply cloaks the card and ignores which permanent it became.
func (c CloakTopOfLibrary) Apply(ctx *Context) error {
	_, err := c.Do(ctx)
	return err
}
