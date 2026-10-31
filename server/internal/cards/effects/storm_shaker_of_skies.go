package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm, Shaker of Skies — {4}{G}{G} Legendary Creature — Mutant Hero 4/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When Storm enters, destroy up to one target artifact, enchantment, or
//	 creature with flying."
//
// The creature clause reads "creature with flying" as a whole: a non-flying
// creature is not a legal target. Choosing no target is legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d936655-8e46-4e30-8f9c-4fa1a4d34a17",
		Name:            "Storm, Shaker of Skies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Storm — destroy up to one target artifact, enchantment, or creature with flying",
					destroyFirstLegalTarget),
				TargetPermanent("up to one target artifact, enchantment, or creature with flying",
					Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))).WithCount(0, 1)),
		},
	})
}
