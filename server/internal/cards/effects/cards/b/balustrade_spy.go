package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balustrade Spy — Creature — Vampire Rogue {3}{B}, 2/3 (Foundations
// Jumpstart):
//
//	"Flying
//	 When this creature enters, target player reveals cards from the top
//	 of their library until they reveal a land card, then puts those
//	 cards into their graveyard."
//
// Shares the mill-until-land body with Undercity Informer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39b4a47a-df51-4bd9-9a20-43ac80a5d352",
		Name:            "Balustrade Spy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Balustrade Spy — target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard",
				effects.TargetPlayerMillsUntilLand),
				effects.TargetPlayer("target player")),
		},
	})
}
