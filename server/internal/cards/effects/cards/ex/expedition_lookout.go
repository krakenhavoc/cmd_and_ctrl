package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expedition Lookout — Creature — Merfolk Rogue {1}{U}, 2/3 (March of the
// Machine, S82 MOM batch 13):
//
//	"Defender
//	 As long as an opponent has eight or more cards in their graveyard,
//	 this creature can attack as though it didn't have defender and it
//	 can't be blocked."
//
// Bristlepack Sentry's conditional defender waiver (CR 702.3b), read live
// at declaration, with the condition "an opponent has eight or more cards
// in their graveyard". The unblockability is a block rule gated on the
// same live read, so it turns on and off with the graveyards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8ad2f88-72ae-4e47-beb5-25a6dacc709b",
		Name:            "Expedition Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as an opponent has eight or more cards in their graveyard, this creature can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return effects.S82OpponentHasGraveyardCards(g, source.Controller, 8)
			})},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return effects.S82OpponentHasGraveyardCards(g, you, 8)
			}),
		},
	})
}
