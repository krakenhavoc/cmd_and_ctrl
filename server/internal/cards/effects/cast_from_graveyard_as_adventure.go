package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cast_from_graveyard_as_adventure.go — "When this creature dies, you may
// cast it from your graveyard as an Adventure until the end of your next
// turn" (Hildibrand Manderville, Mosswood Dreadknight). Append-only.

// castItFromYourGraveyardAsAnAdventure is that dies trigger's body. It
// grants a cast permission over the card in its owner's graveyard,
// narrowed to the Adventure face (the permission's Faces) and good until
// the end of the controller's next turn (UntilEndOfYourNextTurnDuration).
// The grant is the "may": nothing is asked, and an uncast card simply
// stays. It names the object that died (CR 400.7), so a card that left
// the graveyard first is not castable. Cast that way the Adventure
// resolves into exile like any other (CR 715.3d), and the creature can be
// cast from there later.
//
// name is the creature's name and adventure the Adventure's, for the
// permission's attribution.
func castItFromYourGraveyardAsAnAdventure(name, adventure string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		z := g.FindCardZoneForEffect(item.SourceCardID)
		if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     item.Controller,
			Zone:       game.ZoneGraveyard,
			Duration:   g.UntilEndOfYourNextTurnDuration(item.Controller),
			CastOnly:   true,
			Faces:      []int{1},
			Source:     item.SourceCardID,
			SourceName: name,
			Label:      name + " — cast " + adventure + " from your graveyard",
		}, []game.Card{c})
		return nil
	}
}
