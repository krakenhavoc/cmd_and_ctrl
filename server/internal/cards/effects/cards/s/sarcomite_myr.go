package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarcomite Myr — Artifact Creature — Phyrexian Myr {2}{U}, 2/1
// (Future Sight, #5594):
//
//	"{2}: This creature gains flying until end of turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Two activated abilities from shared bodies: the self keyword grant until
// end of turn and Heart Warden's sacrifice-to-draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e78fd8f6-eca9-4107-aee7-a65689cee2f4",
		Name:         "Sarcomite Myr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Sarcomite Myr — flying until end of turn"),
			},
			{
				Label:   "{2}, Sacrifice this creature: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.EbDrawACard,
			},
		},
	})
}
