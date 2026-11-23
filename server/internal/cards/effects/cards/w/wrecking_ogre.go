package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrecking Ogre — Creature — Ogre Warrior {4}{R}, 3/3 (Gatecrash, #5565):
//
//	"Double strike
//	 Bloodrush — {3}{R}{R}, Discard this card: Target attacking creature gets
//	 +3/+3 and gains double strike until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b). Double strike is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85d03947-a39a-4f4f-97c3-6d1094c9248e",
		Name:            "Wrecking Ogre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated:       []effects.ActivatedAbility{effects.GtcBloodrush("Wrecking Ogre", "{3}{R}{R}", 3, 3, "double strike")},
	})
}
