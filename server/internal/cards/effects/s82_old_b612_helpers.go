package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB612RevealTopCreatureElseGraveyard is "Reveal the top card of your library.
// If it's a creature card, put it onto the battlefield. Otherwise, put it into
// your graveyard" (Call of the Wild): a creature goes onto the battlefield under
// the controller's control, anything else into their graveyard as an ordinary
// move, not a mill. A revealed token is neither put nor moved (CR 111.8); an
// empty library does nothing.
func OldB612RevealTopCreatureElseGraveyard(g *game.Game, item *game.StackItem) error {
	top, err := RevealTopThenPutIfMatch(g, item.SourceCardID, item.Controller, game.Card.IsCreature, false,
		"Call of the Wild — revealed from the top of the library")
	if err != nil || top == uuid.Nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(top); !ok || c.IsCreature() || IsToken(c) {
		return nil
	}
	return RestIntoGraveyard(g, []uuid.UUID{top})
}
