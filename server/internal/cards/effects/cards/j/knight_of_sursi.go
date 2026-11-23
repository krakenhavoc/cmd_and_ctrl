package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Sursi — Creature — Human Knight {3}{W}, 2/2 (Future Sight, #5594):
//
//	"Flying; flanking (Whenever a creature without flanking blocks this
//	 creature, the blocking creature gets -1/-1 until end of turn.)
//	 Suspend 3—{W} (Rather than cast this card from your hand, you may
//	 pay {W} and exile it with three time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// Two printed keywords and suspend (CR 702.62a). Flanking's trigger and
// the whole suspend countdown are the engine's, grown from the
// declarations, as on Duskrider Peregrine and Shadow Rider.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8dfe5fb5-89f1-42be-9038-40d3c71ce59a",
		Name:            "Knight of Sursi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "flanking"},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{W}"),
		},
	})
}
