package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b044_helpers.go — shared pieces for the Strixhaven: School of Mages
// batch 044 (S82, #5502).

// ModeTargetPlayerLosesLife is a modal bullet's "target player loses N life"
// step: the life is lost by the player the bullet targeted, if that target is
// still a player as it resolves (Ebony Charm, Silverquill Command). It is one
// Instruction so a bullet can follow it with a step that waits for a prompt
// the life loss asks (a CR 616 order, CR 608.2c).
func ModeTargetPlayerLosesLife(n int) Instruction {
	return func(ctx *Context) error {
		t, ok := ModeTarget(ctx, ctx.Occurrence())
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
	}
}

// ModeReturnTargetToTheBattlefield is a modal bullet's "return target <card>
// from your graveyard to the battlefield": the graveyard card the bullet
// targeted, if it is still a card as it resolves (Against All Odds,
// Silverquill Command, Confront the Past).
func ModeReturnTargetToTheBattlefield(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
