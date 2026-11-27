package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cornered Market — Enchantment {2}{W} (Mercadian Masques, #5619):
//
//	"Players can't cast spells with the same name as a nontoken permanent.
//	 Players can't play nonbasic lands with the same name as a nontoken
//	 permanent."
//
// Exclusion Ritual's cast restriction, with the name read live off the
// battlefield rather than off an exiled card, plus the land-play twin
// (CR 305.1: playing a land is not a cast, so it has its own gate).
// Every player is bound, the controller included; a token does not count
// and a basic land may always be played. Both lift the moment the
// enchantment leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1784bbaa-67b1-4bd1-bf77-53da2bfbf7fc",
		Name:         "Cornered Market",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Players can't cast spells with the same name as a nontoken permanent.",
			Forbids: func(q game.CastQuery) bool {
				return effects.Mmq695SharesNameWithANontokenPermanent(q.Game, q.Card)
			},
		}},
		LandPlayRestrictions: []game.LandPlayRestriction{{
			Label: "Players can't play nonbasic lands with the same name as a nontoken permanent.",
			Forbids: func(q game.LandPlayQuery) bool {
				return !effects.IsBasicLand(q.Card) && effects.Mmq695SharesNameWithANontokenPermanent(q.Game, q.Card)
			},
		}},
	})
}
