package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blacklance Paragon — Creature — Human Knight {1}{B}, 3/1 (Throne of Eldraine, #5516):
//
//	"Flash
//	 When this creature enters, target Knight gains deathtouch and lifelink until end of turn."
//
// Any Knight, either side of the table, itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "258a78cc-a9f3-459a-9837-93a9856f72db",
		Name:            "Blacklance Paragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Blacklance Paragon — target Knight gains deathtouch and lifelink until end of turn",
					effects.JmpB7FirstLegalTargetGains("Blacklance Paragon — deathtouch and lifelink", "deathtouch", "lifelink")),
				effects.TargetCreature("target Knight", effects.HasSubtype("Knight"))),
		},
	})
}
