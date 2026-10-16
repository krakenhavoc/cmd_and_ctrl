package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudsteel Kirin — Artifact Creature — Equipment Kirin {2}{W}, 3/2:
//
//	"Flying
//	 Equipped creature has flying and "You can't lose the game and your
//	 opponents can't win the game."
//	 Reconfigure {5}"
//
// The quoted ability is Platinum Angel's, GIVEN to the equipped
// creature (CR 113.10), so its "you" is that creature's controller
// (CR 109.5), who need not be the Kirin's. It is a granted bundle
// carrying the two gates (#3862, ADR 0093 amendment 2026-10-15): the
// creature has the ability, the loss and win checks read it off the
// creature's composite key like a printed Platinum Angel's, and the
// grant is a layer-6 effect at the Kirin's timestamp, which is the
// moment it became attached (CR 613.7e). So a creature that lost all
// its abilities BEFORE the Kirin was attached has the quoted ability
// (CR 613.6, timestamp order), and one that loses them afterwards does
// not. Flying is an ordinary layer-6 grant beside it.
//
// Reconfigure is reconfigure.go (#2639). No simplification.
func init() {
	const cantLoseKey = "cloudsteel-kirin/cant-lose"
	Register(Spec{
		OracleID:        "a34a71eb-26a4-4e37-8d3c-a0352bea491b",
		Name:            "Cloudsteel Kirin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:          cantLoseKey,
			GameEndGates: YouCantLoseOpponentsCantWin(),
			Text:         "You can't lose the game and your opponents can't win the game.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
			GrantAbilitiesToAttached(cantLoseKey),
		},
		Activated: Reconfigure("{5}"),
	})
}
