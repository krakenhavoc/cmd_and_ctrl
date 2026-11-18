package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arvad the Cursed — Legendary Creature — Vampire Knight {3}{W}{B}, 3/3:
//
//	"Deathtouch, lifelink
//	 Other legendary creatures you control get +2/+2."
//
// The layer 7c anchor is the other-legendary-creatures-you-control
// predicate Shanid uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d561dcff-0a19-46ac-9dee-f374c840d532",
		Name:            "Arvad the Cursed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Static: []game.StaticAbility{
			b16Anthem(otherLegendaryCreaturesYouControl, 2, 2),
		},
	})
}
