package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teleportation Circle — Enchantment {3}{W} (EDHREC top 1,000, #3654):
//
//	"At the beginning of your end step, exile up to one target artifact
//	 or creature you control, then return that card to the battlefield
//	 under its owner's control."
//
// Conjurer's Closet without the "you may": the "up to one" lives on the
// target clause (WithCount(0, 1)) and a trigger with no target is a
// no-op (flickerFirstLegalTarget). The card returns as a new object, so
// its enters triggers fire again (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c786a0aa-d86f-42c3-a7ea-5cb6ab72b5ec",
		Name:         "Teleportation Circle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Teleportation Circle — exile up to one target artifact or creature you control, then return it",
					flickerFirstLegalTarget),
				TargetPermanent("up to one target artifact or creature you control", Or(Artifact(), Creature()), YouControl()).WithCount(0, 1)),
		},
	})
}
