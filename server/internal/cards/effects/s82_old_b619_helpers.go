package effects

import (
	"github.com/google/uuid"
)

// s82_old_b619_helpers.go — S82 OLD batch 619 (#5607): Eighth Edition.

// S82OldB619TapAllMatching is "tap all <permanents matching match>" as one
// instruction: every untapped permanent that matches as the effect resolves is
// tapped. The set is read first and tapped after, so the tap events do not
// disturb the walk (Naya Charm's shape), and the matches are judged through the
// layered view.
func S82OldB619TapAllMatching(ctx *Context, match CardPredicate) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.Tapped && match(ctx.Game, ctx.Controller(), c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
