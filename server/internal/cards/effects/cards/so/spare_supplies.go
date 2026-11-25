package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spare Supplies — Artifact {2} (Zendikar Rising, S82 OLD #5508):
//
//	"This artifact enters tapped.
//	 When this artifact enters, draw a card.
//	 {2}, {T}, Sacrifice this artifact: Draw a card."
//
// Enters tapped is the self-replacement; the sacrifice is a cost
// (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "088fb5dc-65eb-42bf-8a32-76fa6b79b79f",
		Name:         "Spare Supplies",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spare Supplies — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
