package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "f4b1bbd1-f1b2-4c7d-bafd-9d2857f92a89",
		Name:            "Ice-Fang Coatl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.S82GapB22WhileYouControlOtherSnow(3), "deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ice-Fang Coatl — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
