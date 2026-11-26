package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonewright — {R} Creature — Human Shaman 1/1: "Soulbond. As long as
// Stonewright is paired with another creature, each of those creatures has
// "{R}: This creature gets +1/+0 until end of turn.""
//
// No simplification.
func init() {
	const grant = "stonewright/pump"
	const text = "{R}: This creature gets +1/+0 until end of turn."
	effects.Register(effects.Spec{
		OracleID:        "2fe2e0c4-6e02-4f5a-b2b2-920bf912a2bd",
		Name:            "Stonewright",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Stonewright"),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Stonewright — +1/+0"),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.PairedCreaturesHave("As long as Stonewright is paired with another creature, each of those creatures has \""+text+"\"", grant),
		},
	})
}
