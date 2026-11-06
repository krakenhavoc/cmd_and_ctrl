package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flock Impostor — Creature — Shapeshifter {2}{W}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 Flash
//	 Flying
//	 When this creature enters, return up to one other target creature you
//	 control to its owner's hand."
//
// Matterbending Mage's entry bounce limited to your own creatures: "other"
// is object identity (Another), "up to one" lets it target nothing, and a
// target that left in response is skipped. A creature you control that
// someone else owns goes to its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6522070b-7002-44fd-8deb-832fe67ef9a9",
		Name:            "Flock Impostor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Flock Impostor — return up to one other target creature you control to its owner's hand", bounceChosenTarget),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
		},
	})
}
