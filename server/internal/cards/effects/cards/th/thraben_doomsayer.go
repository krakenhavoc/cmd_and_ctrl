package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thraben Doomsayer — {1}{W}{W} Creature — Human Cleric 2/2 (Commander 2020):
//
//	"{T}: Create a 1/1 white Human creature token.
//	 Fateful hour — As long as you have 5 or less life, other creatures you
//	 control get +2/+2."
//
// Fateful hour is an ability word, so the second line is a plain static:
// Lyra's anthem over the other creatures its controller has, switched on while
// that player's life total is 5 or less (DependsOnLifeTotal re-reads the layer
// pass whenever a life total moves).
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.OtherCreaturesYouControl(target, g, source) && effects.YourLifeTotal(g, source) <= 5
	}, 2, 2)
	anthem.Label = "Fateful hour — As long as you have 5 or less life, other creatures you control get +2/+2."
	anthem.DependsOnLifeTotal = true
	effects.Register(effects.Spec{
		OracleID:     "e05e70cd-7af3-428e-82e3-73418914d727",
		Name:         "Thraben Doomsayer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create a 1/1 white Human creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1}),
		}},
		Static: []game.StaticAbility{anthem},
	})
}
