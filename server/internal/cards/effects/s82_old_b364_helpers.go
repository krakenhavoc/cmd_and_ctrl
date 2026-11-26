package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b364_helpers.go — S82 OLD batch 364 (Conspiracy, #5557).

// s82OldB364DestroyBody is the delayed trigger of "destroy it at end of
// combat" (Cinder Wall): it destroys the delayed trigger's source, the
// permanent that blocked, unless it has left the battlefield and come back as
// a new object meanwhile (CR 400.7). A registered key (ADR 0041 phase 3).
var s82OldB364DestroyBody = game.SimpleDelayedBody("s82-old-b364/destroy-at-end-of-combat", func(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	return DestroyTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
})

// S82OldB364DestroyThisAtEndOfCombat is the body of "When this creature
// blocks, destroy it at end of combat": it schedules the destruction for the
// beginning of this combat's end of combat step (CR 511.2).
func S82OldB364DestroyThisAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Body:  s82OldB364DestroyBody,
		}.Apply(NewContext(g, item))
	}
}
