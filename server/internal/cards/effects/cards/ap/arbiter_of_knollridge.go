package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arbiter of Knollridge — Creature — Giant Wizard {6}{W}, 5/5 (Commander 2015,
// S82 OLD #5546):
//
//	"Vigilance
//	 When this creature enters, each player's life total becomes the highest
//	 life total among all players."
//
// The highest total is read once, as the trigger resolves, among the players
// still in the game; each player then gains the difference (CR 119.5), so a
// lifegain trigger sees it, and a player already at the top changes nothing
// (Sorin Markov's B31LifeBecomes).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62705a41-7c4d-4e6c-85b0-36a9d27ab5ae",
		Name:            "Arbiter of Knollridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arbiter of Knollridge — each player's life total becomes the highest life total among all players",
				effects.Do(effects.Instruction(arbiterOfKnollridgeEqualise))),
		},
	})
}

func arbiterOfKnollridgeEqualise(ctx *effects.Context) error {
	best := 0
	first := true
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if first || p.Life > best {
			best, first = p.Life, false
		}
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := effects.B31LifeBecomes(ctx, p.ID, best); err != nil {
			return err
		}
	}
	return nil
}
