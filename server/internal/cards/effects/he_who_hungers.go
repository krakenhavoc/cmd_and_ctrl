package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// He Who Hungers — Legendary Creature — Spirit {4}{B}, 3/2:
//
//	"Flying
//	 {1}, Sacrifice a Spirit: Target opponent reveals their hand. You
//	 choose a card from it. That player discards that card. Activate
//	 only as a sorcery.
//	 Soulshift 4"
//
// The revealed-hand pick (ADR 0116) with no filter, chosen by the
// activator (CR 113.8), at sorcery timing (CR 602.5d). "A Spirit" is
// any permanent you control with the Spirit subtype, He Who Hungers
// itself included. Soulshift 4 is the shared keyword trigger
// (soulshift.go, CR 702.46a), so sacrificing a Spirit to its own
// ability can bring back a cheaper Spirit when He Who Hungers dies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5449bf36-e26c-4f76-bcae-944c70d184f8",
		Name:            "He Who Hungers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{1}, Sacrifice a Spirit: Target opponent reveals their hand. You choose a card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), SacrificeN(1, "a Spirit", HasSubtype("Spirit"))),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect:       TargetRevealsYouChooseDiscardAbility(nil, "card"),
		}},
		Triggered: []game.TriggeredAbility{Soulshift(4, "He Who Hungers")},
	})
}
