package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b270_helpers.go — shared seams of Conspiracy: Take the Crown,
// S82 older-sets batch 270 (#5540).

// ReturnTheFirstClauseTargetToBattlefield is "return target <card> from your
// graveyard to the battlefield" for a trigger with one target clause
// (Eddie Brock, Custodi Soulcaller): the still-legal pick returns under its
// owner's control, and a pick that is no longer legal does nothing (CR 608.2b).
func ReturnTheFirstClauseTargetToBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
