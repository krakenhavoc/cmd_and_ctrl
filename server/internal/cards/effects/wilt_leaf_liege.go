package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:          "2441696b-a9ba-4813-ba2e-e71f85281d05",
		Name:              "Wilt-Leaf Liege",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		Static: []game.StaticAbility{
			b16Anthem(g2OtherCreaturesYouControlOfColor("G"), 1, 1),
			b16Anthem(g2OtherCreaturesYouControlOfColor("W"), 1, 1),
		},
		Replacements: []game.ReplacementEffect{
			OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Wilt-Leaf Liege", nil),
		},
	})
}
