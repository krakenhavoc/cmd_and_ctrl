package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clockwork Fox — Artifact Creature — Fox {3}, 3/2 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"When this creature leaves the battlefield, you draw two cards and each
//	 opponent draws a card."
//
// Any exit, not only dying (CR 603.6c). You draw two, then each opponent draws
// one, each opponent in turn; a player with an empty library still "draws" and
// loses at the next state check (CR 704.5b), so none is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb73db77-4050-4617-a29b-f9a8c885b424",
		Name:         "Clockwork Fox",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Clockwork Fox — you draw two cards and each opponent draws a card", clockworkFoxLeaves),
		},
	})
}

func clockworkFoxLeaves(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	var err error
	effects.EachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
		err = effects.DrawCards{Player: opp, N: 1}.Apply(ctx)
		return err != nil
	})
	return err
}
