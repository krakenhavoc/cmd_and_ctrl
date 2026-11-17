package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandsteppe War Riders — Creature — Human Warrior {3}{G}, 4/4 (March of
// the Machine Commander, #5477):
//
//	"Trample
//	 At the beginning of combat on your turn, bolster X, where X is the
//	 number of differently named artifact tokens you control."
//
// Trample is printed. X is the count of distinct names among the artifact
// tokens you control, read through the layers as the ability resolves
// (s82MomB14DifferentlyNamedArtifactTokens), so two Treasures count once
// and a Treasure and a Food count twice. Bolster (CR 701.39a) chooses the
// creature you control with the least toughness at resolution, with no
// target; X of zero puts nothing anywhere.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "461c0c37-7213-4ece-91f7-934e6a1a8d64",
		Name:            "Sandsteppe War Riders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Sandsteppe War Riders — bolster X, where X is the number of differently named artifact tokens you control",
				Do(s82MomB14BolsterByArtifactTokenNames{})),
		},
	})
}
