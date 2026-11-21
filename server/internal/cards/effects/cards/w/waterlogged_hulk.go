package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waterlogged Hulk // Watertight Gondola — a transforming artifact
// (#2124, ADR 0137):
//
//	Waterlogged Hulk — Artifact {U}
//	  "{T}: Mill a card.
//	   Craft with Island {3}{U}"
//	Watertight Gondola — Artifact — Vehicle, 4/4
//	  "Vigilance
//	   Descend 8 — This Vehicle can't be blocked as long as there are
//	   eight or more permanent cards in your graveyard.
//	   Crew 1"
//
// Craft with Island is a subtype material (CR 702.167b): an Island you
// control — basic or not, or any land an effect made an Island — or an
// Island card in your graveyard. Descend 8 is not a keyword to the
// engine, only Starving Revenant's count of permanent cards in your
// graveyard, read whenever a block is checked (Thieves' Tools'
// can't-be-blocked-while shape).
//
// No simplification.
const WaterloggedHulkOracleID = "30820f71-9fd4-453c-b377-c2e0b01c07a7"

func init() {
	effects.Register(effects.Spec{
		OracleID:     WaterloggedHulkOracleID,
		Name:         "Waterlogged Hulk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Mill a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.MillCards{N: 1}),
			},
			effects.Craft("Craft with Island {3}{U}", "{3}{U}", effects.CraftWithSubtype("Island")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        WaterloggedHulkOracleID + "#1",
		Name:            "Watertight Gondola",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		BlockRules:      []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), descendEight)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Watertight Gondola"),
		}},
	})
}

// descendEight is "as long as there are eight or more permanent cards
// in your graveyard", for the permanent's controller.
func descendEight(g *game.Game, you uuid.UUID, _ game.Card) bool {
	return effects.PermanentCardsInGraveyard(g, you) >= 8
}
