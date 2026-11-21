package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ice-Fang Coatl — Snow Creature — Snake {G}{U}, 1/1 (Mystery Booster 2):
//
//	"Flash
//	 Flying
//	 When this creature enters, draw a card.
//	 This creature has deathtouch as long as you control at least three other
//	 snow permanents."
//
// Flash and flying are printed keywords, the cantrip is a plain enters
// trigger, and the deathtouch is a self-only layer 6 grant read live: the
// Coatl itself is a snow permanent but is never one of the "other" three, and
// the grant ends the moment the count drops below three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f4b1bbd1-f1b2-4c7d-bafd-9d2857f92a89",
		Name:            "Ice-Fang Coatl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{
			KeywordGrant(s82GapB22WhileYouControlOtherSnow(3), "deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ice-Fang Coatl — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
