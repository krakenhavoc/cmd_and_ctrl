package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lyev Skyknight — Creature — Human Knight {1}{W}{U}, 3/1 (Return to
// Ravnica, #5566):
//
//	"Flying
//	 When this creature enters, detain target nonland permanent an
//	 opponent controls. (Until your next turn, that permanent can't attack
//	 or block and its activated abilities can't be activated.)"
//
// Azorius Arrester's trigger widened to any nonland permanent: an artifact
// or planeswalker loses its activated abilities, mana and loyalty ones
// included, until the controller's next turn (CR 701.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "191ec087-7817-4c56-b2d5-66de5c663536",
		Name:            "Lyev Skyknight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Lyev Skyknight — detain target nonland permanent an opponent controls",
				detainEachLegalTarget("Lyev Skyknight — detain")),
				TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls())),
		},
	})
}
