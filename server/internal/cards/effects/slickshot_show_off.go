package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slickshot Show-Off — Creature — Bird Wizard {1}{R}, 1/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying, haste
//	 Whenever you cast a noncreature spell, this creature gets +2/+0
//	 until end of turn.
//	 Plot {1}{R} (You may pay {1}{R} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// Crackling Cyclops' cast trigger at +2/+0, and plot as the ordinary
// special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "74c7a029-59aa-4511-8e4c-f8f167cda4e2",
		Name:            "Slickshot Show-Off",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Slickshot Show-Off — gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Slickshot Show-Off — +2/+0")),
		},
	})
}
