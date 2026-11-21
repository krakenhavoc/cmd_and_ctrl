package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Barricade — Artifact Creature — Wall {1}{W}, 0/4 (Foundations,
// #4465):
//
//	"Defender
//	 You have hexproof.
//	 Prevent all noncombat damage that would be dealt to other creatures
//	 you control."
//
// Defender rides PrintedKeywords and "you have hexproof" is Keen-Eared
// Sentry's player keyword. The prevention is Blessed Sanctuary's
// noncombat static, narrowed to the other creatures you control
// (Vigor's ToAnotherCreatureYouControl): damage to the Barricade itself
// is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "291c7abf-9d64-4ba0-ae21-e3459edc9a95",
		Name:            "Crystal Barricade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		PlayerKeywords:  []string{effects.KeywordHexproof},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToAnotherCreatureYouControl,
				Damage: effects.NoncombatDamage,
				Label:  "Crystal Barricade — prevent all noncombat damage that would be dealt to other creatures you control",
			}),
		},
	})
}
