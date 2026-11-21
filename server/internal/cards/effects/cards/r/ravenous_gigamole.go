package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Gigamole — Creature — Mole Horror {3}{B}, 2/3 (The Brothers' War):
//
//	"When this creature enters, mill three cards. You may put a creature card from
//	 among the cards milled this way into your hand. If you don't, put a
//	 +1/+1 counter on this creature."
//
// Ainok Wayfarer's mill-then-take shape (bro6MillThreeTakeOne), offering
// a creature card from the cards this mill put into the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f4e3e5-2f0d-4d2d-84f6-d1e824a0dcf0",
		Name:         "Ravenous Gigamole",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bro6MillThreeTakeOne("Ravenous Gigamole", "ravenous-gigamole", "a creature card", effects.Creature()),
		},
	})
}
