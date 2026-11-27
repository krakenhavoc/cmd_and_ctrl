package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB629UpkeepIfYouControlEither is "At the beginning of your upkeep, if you
// control a <colour a> or <colour b> permanent, <effect>" (Ceta Sanctuary, S82
// OLD batch 629, #5614), with the CR 603.4 intervening if read as the trigger
// is put on the stack and again as it resolves. The effect is told whether the
// controller also has both colours as it resolves, for the "if you control a
// <a> permanent and a <b> permanent, instead" rider.
func OldB629UpkeepIfYouControlEither(letterA, letterB, name, label string, effect func(g *game.Game, item *game.StackItem, both bool) error) game.TriggeredAbility {
	either := func(g *game.Game, you uuid.UUID) bool {
		return B18ControlsPermanentOfColor(g, you, letterA) || B18ControlsPermanentOfColor(g, you, letterB)
	}
	return InterveningIfYou(name, either, AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		both := B18ControlsPermanentOfColor(g, item.Controller, letterA) && B18ControlsPermanentOfColor(g, item.Controller, letterB)
		return effect(g, item, both)
	}))
}

// OldB629BasicLandTypes are the five basic land types (CR 305.6) Gaea's Balance
// searches for, one land card of each.
var OldB629BasicLandTypes = []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}

// OldB629OneLandOfEachBasicType is the search's set rule: the picks can be
// matched to distinct slots among the five basic land types, each pick having
// that subtype (a land with two of them fills only one slot). An on-disk key:
// never renamed, never reused.
var OldB629OneLandOfEachBasicType = game.RegisterSearchRule("search/s82-old-b629-land-of-each-basic-type", func(picked []game.Card, _ int) bool {
	if len(picked) > len(OldB629BasicLandTypes) {
		return false
	}
	var fill func(i int, used int) bool
	fill = func(i int, used int) bool {
		if i == len(picked) {
			return true
		}
		for slot, t := range OldB629BasicLandTypes {
			if used&(1<<slot) == 0 && picked[i].HasSubtypeWord(t) && fill(i+1, used|1<<slot) {
				return true
			}
		}
		return false
	}
	return fill(0, 0)
})

// OldB629TapTwoUntappedCreaturesCost is "Tap two untapped creatures you
// control" as an ability cost (Symbiotic Deployment). It is not {T}, so
// summoning sickness does not stop a creature paying it (CR 302.6).
func OldB629TapTwoUntappedCreaturesCost() game.AbilityCost {
	const label = "two untapped creatures you control"
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  2,
		Filter: TargetCreature(label),
		Label:  label,
	}}
}
