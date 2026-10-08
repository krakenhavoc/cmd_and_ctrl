package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gastal Raider — Creature — Vampire Rogue {2}{B}, 2/1:
//
//	"Start your engines!
//	 When this creature enters, target opponent reveals their hand. You
//	 choose an instant or sorcery card from it. That player discards
//	 that card.
//	 Max speed — This creature gets +1/+1 and has menace."
//
// ADR 0138 (#2122). The keyword gives its controller speed 1 at the
// next state-based check (CR 702.179a); the engine raises it from there
// (CR 702.179d). The enters trigger is ADR 0116's revealed-hand pick,
// filtered to instants and sorceries. Max speed is a layer 7c pump and
// a layer 6 grant, each switched on by the controller's speed
// (CR 702.178a, effects/speed.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26c09127-8225-4ba2-9f97-f5558b1c41b5",
		Name:            "Gastal Raider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gastal Raider — target opponent reveals their hand, you choose an instant or sorcery card",
					TargetRevealsYouChooseDiscardAbility(Or(Instant(), Sorcery()), "instant or sorcery card")),
				TargetPlayer("target opponent", Opponent())),
		},
		Static: append([]game.StaticAbility{MaxSpeedSelfPump(1, 1)}, MaxSpeedSelfKeywords("menace")...),
	})
}
