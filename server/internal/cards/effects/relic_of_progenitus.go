package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relic of Progenitus — Artifact {1}:
//
//	"{T}: Target player exiles a card from their graveyard.
//	 {1}, Exile this artifact: Exile all graveyards. Draw a card."
//
// The first ability targets a player, and THAT player chooses the card,
// from their own graveyard as the ability resolves (an empty graveyard
// exiles nothing). The second pays {1} and exiles the Relic as a cost,
// then exiles every seat's graveyard (yours too) and draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a9c3401-570c-41cb-a605-75dfa57d6ab7",
		Name:         "Relic of Progenitus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Target player exiles a card from their graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    TapCost(),
				Targets: TargetPlayer("target player"),
				Effect:  rprB15TargetExilesOwnGraveyardCard,
			},
			{
				Label:   "{1}, Exile this artifact: Exile all graveyards. Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard, Draws: 1},
				Cost:    Plus(ManaCost("{1}"), ExileThis()),
				Effect:  rprB15ExileAllGraveyardsThenDraw,
			},
		},
	})
}
