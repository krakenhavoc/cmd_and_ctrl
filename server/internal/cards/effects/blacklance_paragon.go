package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blacklance Paragon — Creature — Human Knight {1}{B}, 3/1 (Throne of Eldraine, #5516):
//
//	"Flash
//	 When this creature enters, target Knight gains deathtouch and lifelink until end of turn."
//
// Any Knight, either side of the table, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "258a78cc-a9f3-459a-9837-93a9856f72db",
		Name:            "Blacklance Paragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Blacklance Paragon — target Knight gains deathtouch and lifelink until end of turn",
					jmpB7FirstLegalTargetGains("Blacklance Paragon — deathtouch and lifelink", "deathtouch", "lifelink")),
				TargetCreature("target Knight", HasSubtype("Knight"))),
		},
	})
}
