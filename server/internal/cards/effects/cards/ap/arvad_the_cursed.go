package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "d561dcff-0a19-46ac-9dee-f374c840d532",
		Name:            "Arvad the Cursed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.OtherLegendaryCreaturesYouControl, 2, 2),
		},
	})
}
