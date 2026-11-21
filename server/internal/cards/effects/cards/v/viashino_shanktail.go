package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Shanktail — Creature — Lizard Warrior {3}{R}, 3/1 (Gatecrash, #5565):
//
//	"First strike
//	 Bloodrush — {2}{R}, Discard this card: Target attacking creature gets
//	 +3/+1 and gains first strike until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b). First strike is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e49f33b-84c0-4b68-abd4-e8c00b3ce98f",
		Name:            "Viashino Shanktail",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.GtcBloodrush("Viashino Shanktail", "{2}{R}", 3, 1, "first strike")},
	})
}
