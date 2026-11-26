package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galvanic Alchemist — {2}{U} Creature — Human Wizard 1/4: "Soulbond. As long as
// this creature is paired with another creature, each of those creatures has
// "{2}{U}: Untap this creature.""
//
// No simplification.
func init() {
	const grant = "galvanic-alchemist/untap"
	const text = "{2}{U}: Untap this creature."
	effects.Register(effects.Spec{
		OracleID:        "694b78c6-443a-4044-84ea-46ac85c8d2cd",
		Name:            "Galvanic Alchemist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Galvanic Alchemist"),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				UntapSelf: true,
				Label:     text,
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.ManaCost("{2}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.PairedCreaturesHave("As long as this creature is paired with another creature, each of those creatures has \""+text+"\"", grant),
		},
	})
}
