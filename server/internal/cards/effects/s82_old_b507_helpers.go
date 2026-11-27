package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b507_helpers.go — Conflux, batch 507 (#5587): the devour that counts
// one creature type apart, and the delayed sacrifice at the next end step.

// OldB507DevourTallyingGoblins is "Devour 1" that also records how many of the
// devoured creatures were Goblins (EntryCardChoice.DevourTally, read back with
// Game.DevouredTallyBy): Voracious Dragon's "twice the number of Goblins it
// devoured".
func OldB507DevourTallyingGoblins(name string) game.ReplacementEffect {
	r := Devour(name, 1)
	r.EntryCardChoice.DevourTally = "Goblin"
	r.RowLabel = "Devour 1: you may sacrifice any number of creatures as it enters; it enters with a +1/+1 counter on it for each."
	return r
}

// OldB507SacrificeThisAtTheNextEndStep is the trigger body of "sacrifice it at
// the beginning of the next end step" (Brackwater Elemental): a delayed trigger
// whose source is this object (CR 603.7d), at the next end step of any
// player's turn, sacrificing it only while it is still the same permanent
// (CR 400.7) and still controlled by the trigger's controller (CR 701.21a).
func OldB507SacrificeThisAtTheNextEndStep(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: label,
			Body:  dorotheaSacrificeBody,
		}.Apply(NewContext(g, item))
	}
}
