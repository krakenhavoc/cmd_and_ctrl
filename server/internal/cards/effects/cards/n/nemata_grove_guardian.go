package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nemata, Grove Guardian — Legendary Creature — Treefolk {4}{G}{G}, 4/5
// (Commander Masters):
//
//	"{2}{G}: Create a 1/1 green Saproling creature token.
//	 Sacrifice a Saproling: Saproling creatures get +1/+1 until end of turn."
//
// The sacrifice is a cost (so it can't be answered by sacrificing the same
// Saproling twice), and the pump reaches the Saproling creatures on the
// battlefield as the ability resolves, an opponent's included, and a
// Saproling that arrives later does not get it (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ce58279-7f1e-4ef7-a352-ecbaee31b6ff",
		Name:         "Nemata, Grove Guardian",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{G}: Create a 1/1 green Saproling creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.ManaCost("{2}{G}"),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
			},
			{
				Label:   "Sacrifice a Saproling: Saproling creatures get +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    effects.SacrificeN(1, "a Saproling", effects.HasSubtype("Saproling")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.HasSubtype("Saproling")), Power: 1, Toughness: 1,
						Label: "Nemata, Grove Guardian — Saproling creatures get +1/+1 until end of turn"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
