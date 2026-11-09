package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amarant Coral — Legendary Creature — Human Monk {2}{R}{G}, 5/4 (Final
// Fantasy, #4462):
//
//	"Trample
//	 Amarant Coral attacks each combat if able.
//	 No Mercy — Whenever Amarant Coral deals combat damage to an opponent,
//	 it deals that much damage to each other opponent."
//
// Hydra Omnivore's trigger: "that much" is the damage that landed on the
// first opponent, read off the triggering event, and the damage to each
// other opponent is not combat damage, so it does not trigger again. The
// requirement to attack is the creature's static (CR 508.1d); a creature
// that can't attack, or whose attack costs something, is not forced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "322d2407-fa9f-48d0-bdf0-88f6d24cf865",
		Name:            "Amarant Coral",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			whenThisHitsAnOpponentThatMuchToEachOther("Amarant Coral — No Mercy: that much damage to each other opponent"),
		},
	})
}
