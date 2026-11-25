package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Time to Reflect — {W} Instant (Amonkhet, #5535):
//
//	"Exile target creature that blocked or was blocked by a Zombie this turn."
//
// The block events of the turn name both sides; the Zombie is read as it
// stands now, or as it was last known if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ce28a3c-5f93-4434-918e-1e74f9fc71d8",
		Name:         "Time to Reflect",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature that blocked or was blocked by a Zombie this turn", timeToReflectCreature),
		OnResolve:    func(item *game.StackItem, ctx *effects.Context) error { return effects.ExileTheFirstTargetIfLegal(ctx) },
	})
}

func timeToReflectCreature(g *game.Game, _ uuid.UUID, c game.Card) bool {
	zombie := func(id uuid.UUID) bool {
		other, ok := g.LookupCardForEffect(id)
		return ok && other.IsCreature() && other.HasSubtypeWord("Zombie")
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventBlock {
			continue
		}
		if (ev.CardID == c.InstanceID && zombie(ev.Target)) || (ev.Target == c.InstanceID && zombie(ev.CardID)) {
			return true
		}
	}
	return false
}
