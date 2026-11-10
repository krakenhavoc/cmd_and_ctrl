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

// CloakTopOfLibrary cloaks the top card of Player's library (the
// controller when zero).
type CloakTopOfLibrary struct {
	Player uuid.UUID
}

// Do cloaks the card and returns the permanent that entered.
func (c CloakTopOfLibrary) Do(ctx *Context) (uuid.UUID, error) {
	player := c.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	id, err := ctx.Game.CloakForEffect(player)
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
