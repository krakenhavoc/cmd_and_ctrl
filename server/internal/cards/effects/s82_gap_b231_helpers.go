package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b231_helpers.go — card-side helpers of set GAP, batch 231
// (Commander Legends: Battle for Baldur's Gate, S82, #5948).

// GreatestOwnedCommanderManaValue is "the greatest mana value of a commander
// you own on the battlefield or in the command zone", zero with none (Majestic
// Genesis's read, exported for the cards that moved out of the core).
func GreatestOwnedCommanderManaValue(g *game.Game, player uuid.UUID) int {
	return greatestOwnedCommanderManaValue(g, player)
}

// ReturnLegalGraveyardTargetsToBattlefieldTogether is "return <the> target
// creature card[s] from your graveyard to the battlefield" as a spell's whole
// resolution (Thwart the Grave): every target still legal as it resolves comes
// back as one event under the caster's control, so each sees the others enter
// (CR 603.6a). With none legal nothing moves.
func ReturnLegalGraveyardTargetsToBattlefieldTogether(ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, ctx.Controller(), false)
}
