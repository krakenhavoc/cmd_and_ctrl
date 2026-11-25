package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wand of Vertebrae — Artifact {1} (Guilds of Ravnica, #5525):
//
//	"{T}: Mill a card.
//	 {2}, {T}, Exile this artifact: Shuffle up to five target cards from your
//	 graveyard into your library."
//
// Excavated Wall's mill, and Tamiyo Meets the Story Circle's shuffle of up to
// five target cards (none included) from your graveyard into your library, with
// Brittle Effigy's exile-this cost. The exile is a cost, so the cards go back
// whatever happens to the Wand afterwards; a target that left the graveyard in
// response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbf7d799-98c9-4dc6-a58c-258bd9c19425",
		Name:         "Wand of Vertebrae",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Mill a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.Do(effects.MillCards{N: 1}),
			},
			{
				Label:   "{2}, {T}, Exile this artifact: Shuffle up to five target cards from your graveyard into your library.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.ExileThis()),
				Targets: effects.TargetCardInGraveyard("up to five target cards from your graveyard", effects.YouOwn()).WithCount(0, 5),
				Effect:  effects.ShuffleTargetsIntoYourLibrary,
			},
		},
	})
}
