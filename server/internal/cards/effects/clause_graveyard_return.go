package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// returnClauseCardFromGraveyard is "return up to one target … card from
// your graveyard to <dest>" for the card that answered clause `slot`,
// when it is still a legal target (CR 608.2b). It returns under its
// owner's control. Shared by Badlands Revival and Geth's Summons.
func returnClauseCardFromGraveyard(ctx *Context, slot int, dest game.ZoneKind) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: dest}.Apply(ctx)
}
