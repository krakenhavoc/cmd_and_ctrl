package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b676_helpers.go — S82 OLD batch 676 (Prophecy, #5617).

// ReturnThisFromYourGraveyardToTheBattlefield is the body of "Return this card
// from your graveyard to the battlefield" (Firemane Angel, Coffin Puppets):
// CR 608.2b, the card may have left the graveyard since the ability was
// activated, and then it does nothing; it returns under its owner's control
// (CR 400.3) as the ability's controller.
func ReturnThisFromYourGraveyardToTheBattlefield(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return (ReturnFromGraveyard{
		Target:     item.SourceCardID,
		Dest:       game.ZoneBattlefield,
		Controller: item.Controller,
	}).Apply(NewContext(g, item))
}
