package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artificer's Dragon — Artifact Creature — Dragon {6}, 4/4:
//
//	"Flying
//	 {R}: Artifact creatures you control get +1/+0 until end of turn.
//	 Unearth {3}{R}{R}"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4284ff9f-4085-4874-abc7-95467cb759a5",
		Name:            "Artificer's Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: Artifact creatures you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Artifact(), effects.Creature(), effects.YouControl()),
						Power: 1,
						Label: "Artificer's Dragon — artifact creatures you control get +1/+0",
					}.Apply(effects.NewContext(g, item))
				},
			},
			effects.Unearth("{3}{R}{R}"),
		},
	})
}
