package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b49_helpers.go — Standard set leftovers (#4452), batches 49 and 50.

// animateLandWithCounters is "put N +1/+1 counters on <land>. It becomes a
// 0/0 <subtype> creature with haste. It's still a land." for a land that is
// still the object it was (Tendril of the Mycotyrant's Fungus, Cosmium
// Confluence's Elemental): awaken's counters-then-animate, with the creature
// type named (game.AnimateLandWithCountersForEffect).
func animateLandWithCounters(ctx *Context, land uuid.UUID, n int, subtype string) error {
	return ctx.Game.AnimateLandWithCountersForEffect(ctx.Controller(), ctx.Source(), land, n, subtype)
}

// cavesYouControl is the battlefield Caves the controller has.
func cavesYouControl(ctx *Context) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(HasSubtype("Cave"), YouControl())) {
		out = append(out, c.InstanceID)
	}
	return out
}

// cosmiumConfluenceAnimate is the pick of "put three +1/+1 counters on a
// Cave you control. It becomes a 0/0 Elemental creature with haste. It's
// still a land.": the controller names the Cave as the mode resolves (it
// does not target), re-checked as it is answered.
var cosmiumConfluenceAnimate = askedPick("choose-cards/cosmium-confluence-animate-cave", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	for _, id := range cavesYouControl(ctx) {
		if id == r.Picked[0] {
			return animateLandWithCounters(ctx, id, 3, "Elemental")
		}
	}
	return nil
})
