package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunted Bonebrute — Creature — Skeleton Beast {2}{B}, 6/2 (Murders at
// Karlov Manor, #4469):
//
//	"Menace
//	 When this creature enters, target opponent creates two 1/1 white Dog
//	 creature tokens.
//	 When this creature dies, each opponent loses 3 life.
//	 Disguise {1}{B} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// The enters trigger is Hunted Horror's: the opponent creates the Dogs,
// so they own and control them (CR 111.2). An opponent who has left by
// the time it resolves gets nothing (CR 608.2b). A face-down Bonebrute
// enters with no abilities (CR 708.2), and turning it face up is not
// entering, so a disguised Bonebrute never hands out Dogs. The dies
// trigger is its controller's: each of their opponents loses 3 life.
// A face-down Bonebrute that dies is a nameless 2/2 with no abilities
// and drains nobody (CR 708.2a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c1d9c41a-0941-4b7e-b8a6-890d30386458",
		Name:             "Hunted Bonebrute",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hunted Bonebrute — target opponent creates two 1/1 white Dog creature tokens", huntedBonebruteDogs),
				TargetPlayer("target opponent", Opponent())),
			WhenThisDies("Hunted Bonebrute — each opponent loses 3 life", Do(EachOpponentLosesLife{N: 3})),
		},
	})
}

// huntedBonebruteDogs gives the targeted opponent two 1/1 white Dogs.
func huntedBonebruteDogs(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return CreateToken{Controller: t.ID, Template: TokenCard("1/1 white Dog"), N: 2}.Apply(ctx)
		}
	}
	return nil
}
