package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Noble — Creature — Faerie Noble {2}{G}, 1/2 (Homelands, S82 OLD b823,
// #5635):
//
//	"Flying
//	Other Faerie creatures you control get +0/+1.
//	{T}: Other Faerie creatures you control get +1/+0 until end of turn."
//
// Scion of Oona's lord filter for the static, and War Screecher's group boost for
// the tap ability: it reaches the Faeries you control as it resolves, not ones
// that arrive later (CR 611.2c).
//
// No simplification.
func init() {
	otherFaeries := effects.TribeFilter{Tribes: []string{"Faerie"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "6172e5b9-2a01-429c-a140-e9b22fbab56e",
		Name:            "Faerie Noble",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.TribalAnthem(otherFaeries, 0, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Other Faerie creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Faerie"), effects.OtherThan(item.SourceCardID)), Power: 1,
					Label: "Faerie Noble — other Faeries you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
