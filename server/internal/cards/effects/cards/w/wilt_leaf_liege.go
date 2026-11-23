package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wilt-Leaf Liege — Creature — Elf Knight {1}{G/W}{G/W}{G/W}, 4/4:
//
//	"Other green creatures you control get +1/+1.
//	 Other white creatures you control get +1/+1.
//	 If a spell or ability an opponent controls causes you to discard
//	 this card, put it onto the battlefield instead of putting it into
//	 your graveyard."
//
// Glen Elendra Liege's two statics (a creature that is both green and
// white gets +2/+2, each ability applying on its own) plus the Baloth
// replacement (ADR 0153, see obstinate_baloth.go). The hybrid cost is
// the printed {1}{G/W}{G/W}{G/W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "2441696b-a9ba-4813-ba2e-e71f85281d05",
		Name:              "Wilt-Leaf Liege",
		Completeness:      effects.CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Replacements: []game.ReplacementEffect{
			effects.OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Wilt-Leaf Liege", nil),
		},
	})
}
