package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Wind Guide — {2}{W}{U} Creature — Bird Warrior 2/3 (Amonkhet, S82 old
// batch 247):
//
//	"Flying, vigilance
//	 Creature tokens you control have flying and vigilance.
//	 Embalm {4}{W}{U}"
//
// A layer 6 grant (CR 613.1f) to every creature token its controller controls,
// itself included when it is a token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb813ffa-2615-4177-a7fb-6e0dfb819fc2",
		Name:            "Aven Wind Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Embalm("{4}{W}{U}")},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.CreatureTokensYouControl, "flying", "vigilance"),
		},
	})
}
