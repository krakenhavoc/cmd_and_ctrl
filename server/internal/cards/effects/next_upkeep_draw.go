package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// next_upkeep_draw.go — "Draw a card at the beginning of the next turn's
// upkeep." (#5634: Carrier Pigeons, Fevered Strength, Burnout; the same
// clause Urza's Bauble, Foxfire and Gravebind print). It is a CR 603.7a
// delayed trigger scheduled for the very next upkeep at the table,
// whoever's turn that is, controlled by the player who controlled the
// spell or ability that scheduled it. The draw goes on the stack when
// that upkeep begins, so it can be responded to.
//
// Append-only: add a builder, never change what one means.

// drawACardAtTheNextUpkeep is the clause as a primitive, so a spell
// applies it after its other instructions and a trigger sequences it
// with Do. `label` names the delayed trigger ("<Card> — draw a card").
func drawACardAtTheNextUpkeep(label string) ScheduleDelayedTrigger {
	return ScheduleDelayedTrigger{
		At:    game.StepUpkeep,
		Label: label,
		Body:  drawOneBody,
	}
}
