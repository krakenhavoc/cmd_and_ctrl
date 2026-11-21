package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur Devotee — Creature — Human Druid {1}{U}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Defender
//	 {1}: Add {G}, {U}, or {R}. Activate only once each turn."
//
// Abzan Devotee's filter with the Temur colours: a Signet-style
// `ManaAbilityCost.Mana` with no tap, limited by the declarative
// OncePerTurn bit, counted per object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e444f072-ce24-4fd1-99eb-d20d12bc119a",
		Name:            "Temur Devotee",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{G|U|R}",
			Label:       "{1}: Add {G}, {U}, or {R}",
			OncePerTurn: true,
		}},
	})
}
