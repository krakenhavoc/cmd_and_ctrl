package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82GapB23VelukanRolled is Velukan Dragon's tail: "this creature gets +X/+0
// until end of turn, where X is the result minus 1." An on-disk key, never
// renamed, never reused. The Dragon must still be the permanent the trigger
// came from (CR 400.7); the result is the die's final one (CR 706.2), so a
// modifier on the roll counts, and a roll that kept no die gives no bonus.
var s82GapB23VelukanRolled = CardTail("dice/velukan-dragon-plus-result-minus-one", func(ctx *Context, o game.TailOutcome) error {
	x := DiceResult(o) - 1
	if x <= 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: x, Label: "Velukan Dragon — +X/+0 until end of turn"}.Apply(ctx)
})

// s82GapB23RecruiterOrder is "put those cards on top in any order": the found
// cards are already on top and the controller arranges them. An on-disk key,
// never renamed, never reused.
var s82GapB23RecruiterOrder = CardTail("then/goblin-recruiter-order", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) < 2 {
		return nil
	}
	return PutInLibraryInAnyOrder{
		Cards:     o.IDs,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Label:     "Goblin Recruiter — put those cards on top in any order",
	}.Apply(ctx)
})
