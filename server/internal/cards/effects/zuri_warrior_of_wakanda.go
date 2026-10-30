package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "aa792980-0811-4c08-8343-00fba0c5cf5e",
		Name:            "Zuri, Warrior of Wakanda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Artifact(), ManaValueGE(4)),
				"Zuri, Warrior of Wakanda — a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}
