package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minwu, White Mage — Legendary Creature — Human Cleric {3}{W}{W}, 3/3
// (Final Fantasy, #4462):
//
//	"Vigilance, lifelink
//	 Whenever you gain life, put a +1/+1 counter on each Cleric you
//	 control."
//
// The life-gain trigger is Biogenic Ooze's "a +1/+1 counter on each
// <type> you control", read as it resolves: Minwu is a Cleric, so she
// gets one too, and her own lifelink damage sets it off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "31ca9e12-8ad0-472e-8b2a-b7df26370d8c",
		Name:            "Minwu, White Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Minwu, White Mage — put a +1/+1 counter on each Cleric you control",
				putAPlusOneCounterOnEachYouControl(Subtype("Cleric"))),
		},
	})
}
