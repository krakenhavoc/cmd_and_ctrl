package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stern Mentor — {3}{U} Creature — Human Wizard 2/2: "Soulbond. As long as this
// creature is paired with another creature, each of those creatures has "{T}:
// Target player mills two cards.""
//
// No simplification.
func init() {
	const grant = "stern-mentor/mill"
	const text = "{T}: Target player mills two cards."
	effects.Register(effects.Spec{
		OracleID:        "6dab0e1a-f39a-4113-9d13-b067cfefbdb3",
		Name:            "Stern Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Stern Mentor"),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.TargetPlayerMillsTwo,
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.PairedCreaturesHave("As long as this creature is paired with another creature, each of those creatures has \""+text+"\"", grant),
		},
	})
}
