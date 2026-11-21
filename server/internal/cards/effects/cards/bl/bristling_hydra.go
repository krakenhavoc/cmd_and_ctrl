package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bristling Hydra — Creature — Hydra {2}{G}{G}, 4/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}{E}{E}: Put a +1/+1 counter on this creature. It gains
//	 hexproof until end of turn."
//
// ADR 0129 PR 1 (#1995). Both halves land on the Hydra that activated
// it; one that left and came back is a new object and gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3b23c58-0b7a-4fe4-a8e8-5320a7605724",
		Name:         "Bristling Hydra",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Bristling Hydra", 3),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}{E}{E}: Put a +1/+1 counter on this creature. It gains hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerProtect},
			Cost:    effects.PayEnergy(3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("hexproof", "Bristling Hydra — hexproof until end of turn")(g, item)
			},
		}},
	})
}
