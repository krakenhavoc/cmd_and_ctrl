package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b688_helpers.go — shared pieces of the Mercadian Masques creatures of batch 688
// (S82 OLD, #5619).

// OldB688SacrificeAMountainOrThisAnswer is the keyed answer to Lithophage's "sacrifice this
// creature unless you sacrifice a Mountain": index 0 asks for a Mountain, any other index
// sacrifices the permanent. An on-disk key: never renamed, never reused.
var OldB688SacrificeAMountainOrThisAnswer = sacrificeOneOrThisAnswer(
	"option-pick/sacrifice-this-unless-you-sacrifice-a-mountain", "a Mountain", HasSubtype("Mountain"))

// oldB688TuckBody is the delayed trigger of "put it on top of its owner's library at end of
// combat": it tucks the delayed trigger's source unless it has left the battlefield or come
// back as a new object meanwhile (CR 400.7). A registered key (ADR 0041 phase 3).
var oldB688TuckBody = game.SimpleDelayedBody("s82-old-b688/tuck-at-end-of-combat", func(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	return g.TuckToLibraryForEffect(item.SourceCardID, false)
})

// OldB688TuckThisAtEndOfCombat is the body of "When this creature attacks or blocks, put it on
// top of its owner's library at end of combat": it schedules the tuck for the beginning of this
// combat's end of combat step (CR 511.2).
func OldB688TuckThisAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Body:  oldB688TuckBody,
		}.Apply(NewContext(g, item))
	}
}
