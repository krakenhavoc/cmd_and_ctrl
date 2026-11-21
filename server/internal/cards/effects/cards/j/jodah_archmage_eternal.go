package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jodah, Archmage Eternal — Legendary Creature — Human Wizard {1}{U}{R}{W}, 4/3:
//
//	"Flying
//	 You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for
//	 spells you cast."
//
// Fist of Suns' static on a flier (ADR 0118 §3, #2163). A Jodah that has
// lost its abilities (Humility) grants nothing, because the grant is
// read from the permanent's catalog abilities.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:                "8be4745e-36d8-430f-945e-c8a7fde9b4f6",
		Name:                    "Jodah, Archmage Eternal",
		Completeness:            effects.CompletenessFull,
		PrintedKeywords:         []string{"flying"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.PayWUBRGForSpellsYouCast()},
	})
}
