package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b491_helpers.go — Zendikar kicker batch 491 (S82 older sets, #5583).

// S82OldB491WhenThisEntersCapturingKicked is an enters trigger that records,
// as it is put on the stack, how many times the permanent was kicked
// (CR 702.33d) in Params.Amount: "if it wasn't kicked" is then answered from
// what the creature was as it entered, not from a card that may have left the
// battlefield since (CR 603.4, CR 400.7d). Sphinx of Lost Truths.
func S82OldB491WhenThisEntersCapturingKicked(label string, effect Effect) game.TriggeredAbility {
	t := WhenThisEnters(label, effect)
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = game.CardKickedTimes(*source)
		return item
	}
	return t
}
