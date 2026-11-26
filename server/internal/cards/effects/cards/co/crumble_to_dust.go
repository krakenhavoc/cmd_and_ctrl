package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crumble to Dust — Sorcery {3}{R} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Exile target nonbasic land. Search its controller's graveyard, hand, and
//	 library for any number of cards with the same name as that land and exile
//	 them. Then that player shuffles."
//
// Sowing Salt's body with "any number" in place of "all"
// (OldB308ExilePermanentThenExtractAnyNumber): the searcher takes or leaves
// every match, in the graveyard too. "Nonbasic" is the Basic supertype, so a
// nonbasic land given a basic land type is still a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c13da0f4-917f-4e07-b484-21cb4bec4da6",
		Name:            "Crumble to Dust",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.TargetPermanent("target nonbasic land", effects.B10NonbasicLand()),
		OnResolve:       effects.OldB308ExilePermanentThenExtractAnyNumber("Crumble to Dust"),
	})
}
