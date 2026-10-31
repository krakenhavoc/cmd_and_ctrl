package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Squadron Sinister — Legendary Creature — Human Villain
// {5}{U}{R}, 5/5:
//
//	"Flying, haste
//	 Other Villains you control get +2/+2 and have flying and haste.
//	 Mayhem {3}{U}{R} (You may cast this card from your graveyard for
//	 {3}{U}{R} if you discarded it this turn. Timing rules still apply.)"
//
// Beorn's tribal lord (TribeFilter with Others and YoursOnly) for the
// pump and one keyword grant per word. Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	villains := TribeFilter{Tribes: []string{"Villain"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:         "3b7352e6-6508-491c-9935-426a6eebb429",
		Name:             "The Squadron Sinister",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "haste"},
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{3}{U}{R}")},
		Static: []game.StaticAbility{
			TribalAnthem(villains, 2, 2),
			TribalKeywordGrant(villains, "flying"),
			TribalKeywordGrant(villains, "haste"),
		},
	})
}
