package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Valor's Reach — Creature — Archon {4}{G}{W}, 5/6 (Battlebond,
// #5529):
//
//	"Flying, vigilance, trample
//	 As this creature enters, choose artifact, enchantment, instant,
//	 sorcery, or planeswalker.
//	 Players can't cast spells of the chosen type."
//
// The keywords ride PrintedKeywords. The choice is Serra's Emissary's CR
// 614.12 option pick over the five printed words, stored on
// Card.ChosenOption (per instance, cleared when the Archon leaves, CR
// 400.7). The ban is a permanent's printed cast restriction
// (Spec.CastRestrictions, bbdCastBanOfTheChosenType), so it lasts exactly
// as long as the Archon is on the battlefield: every player, its controller
// included, is refused a spell with the chosen card type, an artifact
// creature spell included when artifact was chosen. Unanswered, nothing is
// refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "baff00b8-b109-4e31-942b-38aeb1e45ab2",
		Name:             "Archon of Valor's Reach",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "vigilance", "trample"},
		AsEnters:         ChooseOptionAsEnters("Archon of Valor's Reach", "Artifact", "Enchantment", "Instant", "Sorcery", "Planeswalker"),
		CastRestrictions: []game.CastRestriction{bbdCastBanOfTheChosenType("Players can't cast spells of the chosen type.")},
	})
}
