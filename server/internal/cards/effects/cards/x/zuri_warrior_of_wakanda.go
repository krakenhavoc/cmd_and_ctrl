package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuri, Warrior of Wakanda — Legendary Creature — Human Warrior Hero {1}{G},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 Whenever you cast an artifact spell with mana value 4 or greater, put a
//	 +1/+1 counter on each creature you control."
//
// Jin-Gitaxias' cast trigger with an artifact and mana-value filter, and
// Archangel of Thune's counter on each creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa792980-0811-4c08-8343-00fba0c5cf5e",
		Name:            "Zuri, Warrior of Wakanda",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Artifact(), effects.ManaValueGE(4)),
				"Zuri, Warrior of Wakanda — a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl),
		},
	})
}
