package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Wasp, Winsome Avenger — {1}{U} Legendary Creature — Human Hero 2/1
// (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Flying
//	 When The Wasp enters, target Hero gains hexproof until end of turn.
//	 Whenever The Wasp attacks, tap target creature defending player
//	 controls."
//
// The enters target is any Hero on the battlefield, The Wasp included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b706959-ea38-4808-870a-7c35d83f7ab2",
		Name:            "The Wasp, Winsome Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenThisEnters("The Wasp — target Hero gains hexproof until end of turn",
						ebTargetCreatureGainsUntilEOT("hexproof", "The Wasp, Winsome Avenger — hexproof")),
					TargetCreature("target Hero", Subtype("Hero"))),
				ForTargets(BenefitUntilEOT(0, 0, 0, "hexproof"))),
			distractingGeistTap("The Wasp — tap target creature defending player controls"),
		},
	})
}
