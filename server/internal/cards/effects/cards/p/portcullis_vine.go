package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portcullis Vine — Creature — Plant Wall {G}, 0/3 (Core Set 2021, #5511):
//
//	"Defender
//	 {2}, {T}, Sacrifice a creature with defender: Draw a card."
//
// The sacrifice is a cost and may be the Vine itself (it has defender); a
// creature without defender is refused. The keyword is read through the
// layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a217f386-11b5-48f0-b173-3b0700ba295f",
		Name:            "Portcullis Vine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a creature with defender: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeN(1, "a creature with defender", effects.Creature(), effects.HasKeyword("defender"))),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
