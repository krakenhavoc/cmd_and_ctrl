package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benevolent River Spirit — Creature — Spirit {U}{U}, 4/5 (Avatar: The
// Last Airbender, #4459):
//
//	"As an additional cost to cast this spell, waterbend {5}. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures
//	 to help. Each one pays for {1}.)
//	 Flying, ward {2}
//	 When this creature enters, scry 2."
//
// The waterbend is a mandatory additional cost, Waterbender's
// Restoration's Spec.TapCost with a fixed {5}: the caster pays {U}{U}
// plus five, any of the five by tapping an untapped artifact or
// creature they control (CR 701.67a). Ward {2} is the catalog's ward
// trigger, and the entry scry is the footprinted Scry primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12d91743-1ccb-4f2b-b2a9-1c2b34f7c8eb",
		Name:            "Benevolent River Spirit",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 2},
		TapCost:         Waterbend("{5}"),
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Benevolent River Spirit — ward {2}"),
			WhenThisEnters("Benevolent River Spirit — scry 2", Do(Scry{N: 2})),
		},
	})
}
