package effects

import "github.com/google/uuid"

// s82_gap_b131_helpers.go — GAP batch 131 (S82 #5948).

// GapB131ReturnedManaValue is the mana value "equal or lesser" is measured
// against in Press the Enemy: the target's as it is now, read before it is
// returned (CR 608.2h). A spell on the stack counts the X it was cast with
// (CR 202.3e) and a permanent's X is 0, which is what ManaValueForEffect
// reads.
func GapB131ReturnedManaValue(ctx *Context, id uuid.UUID) int {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return 0
	}
	mv, _ := ctx.Game.ManaValueForEffect(c)
	return mv
}
