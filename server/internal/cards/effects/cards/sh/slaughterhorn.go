package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slaughterhorn — Creature — Beast {2}{G}, 3/2 (Gatecrash, #5565):
//
//	"Bloodrush — {G}, Discard this card: Target attacking creature gets +3/+2
//	 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1fca987-3484-4a76-86b6-d61ef0a2cf0f",
		Name:         "Slaughterhorn",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.GtcBloodrush("Slaughterhorn", "{G}", 3, 2, "")},
	})
}
