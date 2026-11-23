package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Airlift Chaplain — Creature — Human Cleric {2}{W}, 1/1 (The Brothers' War):
//
//	"Flying
//	 When this creature enters, mill three cards. You may put a Plains card
//	 or a creature card with mana value 3 or less from among the cards
//	 milled this way into your hand. If you don't, put a +1/+1 counter on
//	 this creature."
//
// Ainok Wayfarer's mill-then-take shape (bro6MillThreeTakeOne), offering a
// Plains card or a creature card with mana value 3 or less from the cards
// this mill put into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "719aeb50-c8cf-4cd8-b74c-d2cef408b52e",
		Name:            "Airlift Chaplain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Bro6MillThreeTakeOne("Airlift Chaplain", "airlift-chaplain", "a Plains card or a creature card with mana value 3 or less",
				effects.Or(effects.HasSubtype("Plains"), effects.And(effects.Creature(), effects.ManaValueLE(3)))),
		},
	})
}
