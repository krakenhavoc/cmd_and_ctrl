package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ReturnTargetedGraveyardCardToBattlefield is "return target <card> from
// your graveyard to the battlefield" as a trigger's whole Effect, the card
// entering untapped: Quarry Beetle. The target is re-checked as the effect
// resolves (CR 608.2b), so a card that left the graveyard is skipped.
func ReturnTargetedGraveyardCardToBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || len(ctx.LegalTargets()) == 0 {
		return nil
	}
	return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
