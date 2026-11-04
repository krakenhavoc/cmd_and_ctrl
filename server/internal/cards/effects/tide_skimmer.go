package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tide Skimmer — Creature — Drake {3}{U}, 2/3 (Jumpstart 2020):
//
//	"Flying
//	 Whenever you attack with two or more creatures with flying, draw a
//	 card."
//
// Spinneret and Spiderling's "attack with two or more" shape: each
// creature with flying you declare as an attacker is one attack event
// (jmpB9YouAttackedWithAFlier), and the ability waits for the
// declaration to finish (AtBatchEnd, ThatManyEventsAtLeast(2)), so it
// triggers once per declaration with two or more fliers, the Skimmer
// itself counting when it attacks. A second combat that declares two
// fliers again triggers it again.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:        "8d83dfb3-5662-4109-b465-bf5f257e58a5",
		Name:            "Tide Skimmer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventAttack, jmpB9YouAttackedWithAFlier,
				"Tide Skimmer — you attacked with two or more creatures with flying: draw a card",
				Do(DrawCards{N: 1})), ThatManyEventsAtLeast(2)),
		},
	})
}
