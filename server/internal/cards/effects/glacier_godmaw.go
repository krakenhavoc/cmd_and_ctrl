package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glacier Godmaw — Creature — Leviathan {5}{G}{G}, 6/6 (Edge of
// Eternities, #4461):
//
//	"Trample
//	 When this creature enters, create a Lander token.
//	 Landfall — Whenever a land you control enters, creatures you control
//	 get +1/+1 and gain vigilance and haste until end of turn."
//
// Galactic Wayfarer's enters Lander (b16LanderToken) and Appa, the
// Vigilant's team pump on a landfall trigger. The creatures it affects
// are fixed as it resolves (CR 611.2c), so a creature that enters later
// in the turn gets nothing from that instance, and each land that
// enters is its own trigger. Trample is a printed keyword.
//
// No simplification.
func init() {
	you := And(Creature(), YouControl())
	Register(Spec{
		OracleID:        "300e48df-7f7f-47ff-b450-8ba7bd1b1c67",
		Name:            "Glacier Godmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glacier Godmaw — create a Lander token",
				Do(CreateToken{Template: b16LanderToken(), N: 1})),
			Landfall("Glacier Godmaw — creatures you control get +1/+1 and gain vigilance and haste until end of turn",
				Do(
					BoostUntilEOT{Match: you, Power: 1, Toughness: 1, Label: "Glacier Godmaw — +1/+1"},
					GrantKeywordUntilEOT{Match: you, Keywords: []string{"vigilance", "haste"}, Label: "Glacier Godmaw — vigilance and haste"},
				)),
		},
	})
}
