package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wasteland Viper — Creature — Snake {G}, 1/2 (Gatecrash, #5565):
//
//	"Deathtouch
//	 Bloodrush — {G}, Discard this card: Target attacking creature gets +1/+2
//	 and gains deathtouch until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b). Deathtouch is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e30b4eb8-6393-486a-9f8e-970ad2351d8a",
		Name:            "Wasteland Viper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []effects.ActivatedAbility{effects.GtcBloodrush("Wasteland Viper", "{G}", 1, 2, "deathtouch")},
	})
}
