package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Namora, the Sea Queen — {2}{U} Legendary Creature — Mutant Merfolk Hero 2/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Power-up — {5}{U}: Put a +1/+1 counter on Namora. Create two 1/1 blue
//	 Merfolk creature tokens. (Activate each power-up ability only once.
//	 Reduce the cost by her mana cost if she entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The tokens are the catalog's
// "1/1 blue Merfolk" row (Namor the Sub-Mariner).
//
// No simplification.
func init() {
	ab := effects.PowerUp("{5}{U}: Put a +1/+1 counter on Namora. Create two 1/1 blue Merfolk creature tokens.",
		effects.ManaCost("{5}{U}"), func(g *game.Game, item *game.StackItem) error {
			if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk"), N: 2}.Apply(effects.NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:        "d0cd3edd-d908-47c5-bff1-2089c22ab41c",
		Name:            "Namora, the Sea Queen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{ab},
	})
}
