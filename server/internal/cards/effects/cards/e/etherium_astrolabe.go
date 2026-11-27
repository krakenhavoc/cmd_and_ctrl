package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Etherium Astrolabe — Artifact {2}{U} (Shards of Alara, S82 OLD batch 509,
// #5588):
//
//	"Flash
//	 {B}, {T}, Sacrifice an artifact: Draw a card."
//
// The artifact may be the Astrolabe itself (SacrificeN, not SacrificeAnotherN).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d7194f7-1525-4f6e-b974-b947a254f437",
		Name:            "Etherium Astrolabe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Sacrifice an artifact: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
