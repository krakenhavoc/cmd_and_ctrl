package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fervor — Enchantment {2}{R} (EDHREC rank 3517):
//
//	"Creatures you control have haste. (They can attack and {T} as
//	 soon as they come under your control.)"
//
// The haste anthem. A layer-6 grant to every creature its controller
// controls — Fervor itself is not a creature and grants itself
// nothing — through the same TribalKeywordGrant a lord's keyword
// half uses, with no tribe named (an empty TribeFilter matches every
// creature). The engine's summoning-sickness gate reads the
// effective keyword, so a creature that entered this turn attacks
// and taps at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e0cea9c-3110-4728-9378-76849e33bb90",
		Name:         "Fervor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "haste"),
		},
	})
}
