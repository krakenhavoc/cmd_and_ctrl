package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zamriel, Seraph of Steel — Legendary Creature — Angel {2}{W}{W}, 3/4
// (Game Night: Free-for-All, #5485):
//
//	"Flying
//	 During your turn, equipped creatures you control have indestructible."
//
// Long-Lost Lances' "during your turn, equipped creatures you control"
// (hobB5EquippedCreatureOnYourTurn) as a layer 6 keyword grant. A static
// ability's effect is not locked in (CR 611.3a): the grant is read on
// every recompute, so it is live only while Zamriel's controller is the
// active player and only on creatures equipped at that moment.
// "Equipped" is an Equipment attached to the creature (CR 301.5a),
// whoever controls the Equipment. Zamriel herself is covered when she is
// equipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "acb0fa6a-89c8-4957-881e-c61659e86f97",
		Name:            "Zamriel, Seraph of Steel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(effects.HobB5EquippedCreatureOnYourTurn, "indestructible"),
				"During your turn, equipped creatures you control have indestructible."),
		},
	})
}
