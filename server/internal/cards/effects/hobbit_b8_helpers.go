package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hobB8ReturnTargetGraveyardCardToBattlefield is "put/return target card
// from a graveyard onto the battlefield" for a card the ability targeted:
// it lands under its owner's control, and nothing happens when the pick
// was declined (an "up to one" clause) or is gone. Bilbo, Unexpected
// Adventurer, Emeria, the Sky Ruin and Goblin Engineer share it.
func hobB8ReturnTargetGraveyardCardToBattlefield(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
}
