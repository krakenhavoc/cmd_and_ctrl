package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Wolves — Creature — Wolf {2}{G}, 2/2 (Ice Age, S82 OLD, #5636):
//
//	"This creature has banding as long as you control a Plains."
//
// A layer 6 keyword grant (CR 613.1f) on itself while its controller controls a Plains, read live, so
// losing the last Plains takes banding away mid-combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d72a2f4-cce6-4a2b-9507-0ee1c3e1e4ae",
		Name:         "Dire Wolves",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				s := effects.KeywordGrant(effects.S82OldB843YouControlAPlains, game.KeywordBanding)
				s.Label = "This creature has banding as long as you control a Plains."
				return s
			}(),
		},
	})
}
