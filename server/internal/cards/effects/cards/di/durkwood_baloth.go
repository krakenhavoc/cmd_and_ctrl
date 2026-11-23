package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Durkwood Baloth — Creature — Beast {4}{G}{G}, 5/5 (Time Spiral, #5596):
//
//	"Suspend 5—{G} (Rather than cast this card from your hand, you may
//	 pay {G} and exile it with five time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// A vanilla 5/5 whose only other text is suspend (CR 702.62a). The keyword
// is the engine's: the special action, the time counters, the upkeep
// countdown, the free cast when the last counter comes off and the haste
// it then has are all grown from the one declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1732cd31-1393-426f-a2fe-9b6b3f27136b",
		Name:         "Durkwood Baloth",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(5, "{G}"),
		},
	})
}
