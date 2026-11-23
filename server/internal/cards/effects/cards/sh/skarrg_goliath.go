package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skarrg Goliath — Creature — Beast {6}{G}{G}, 9/9 (Gatecrash, #5565):
//
//	"Trample
//	 Bloodrush — {5}{G}{G}, Discard this card: Target attacking creature gets
//	 +9/+9 and gains trample until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b). Trample is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72350461-8231-4737-a57f-f20fb5d179a3",
		Name:            "Skarrg Goliath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.GtcBloodrush("Skarrg Goliath", "{5}{G}{G}", 9, 9, "trample")},
	})
}
