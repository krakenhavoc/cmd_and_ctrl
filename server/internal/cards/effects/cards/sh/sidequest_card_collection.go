package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sidequest: Card Collection // Magicked Card (Final Fantasy, #4462).
// Front, Enchantment {3}{U}:
//
//	"When this enchantment enters, draw three cards, then discard two
//	 cards.
//	 At the beginning of your end step, if eight or more cards are in
//	 your graveyard, transform this enchantment."
//
// Back, Artifact — Vehicle, 4/4:
//
//	"Flying
//	 Crew 1"
//
// The enters ability is Captain of Umbar's loot with a bigger shape
// (drawThenDiscard 3 then 2: the discard prompt is built from the
// post-draw hand). The end-step ability is an intervening "if" (CR 603.4,
// finB1CondEightInYourGraveyard): checked as the step begins and again
// as the ability resolves, as the ruling says. The transform is the
// in-place verb on the enchantment; the back face is a Vehicle that
// crews as any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     SidequestCardCollectionOracleID,
		Name:         "Sidequest: Card Collection",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sidequest: Card Collection — draw three cards, then discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 3, 2, "")
				}),
			effects.InterveningIf(effects.FinB1CondEightInYourGraveyard{},
				effects.AtYourEndStep("Sidequest: Card Collection — transform this enchantment (eight or more cards are in your graveyard)",
					effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        SidequestCardCollectionOracleID + "#1",
		Name:            "Magicked Card",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Magicked Card"),
		}},
	})
}

// sidequestCardCollectionOracleID is shared by both faces.
const SidequestCardCollectionOracleID = "0c2095b0-478d-45b4-a34d-b19eb77cea4c"
