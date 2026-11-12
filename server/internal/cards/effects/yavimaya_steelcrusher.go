package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Steelcrusher — Creature — Ape Warrior {1}{R}, 2/2 (Dominaria
// United, #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 {1}, Sacrifice this creature: Destroy target artifact."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// The activated ability is District Mascot's destroy with a sacrifice of
// the Steelcrusher itself as a cost (CR 602.2b); with the target gone at
// resolution it does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad181ee7-756f-4002-9009-023ea6172991",
		Name:            "Yavimaya Steelcrusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
