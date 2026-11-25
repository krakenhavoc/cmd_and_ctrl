package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of the Great Conduit — Creature — Elemental {3}{G}, 2/4 (Kaladesh,
// #5539):
//
//	"Reach
//	 As long as you control a Nissa planeswalker, this creature gets +2/+0
//	 and has vigilance."
//
// One condition gating two statics on the Guardian alone, as Goblin Tomb
// Raider does: a layer 7c +2/+0 (CR 613.4c) and a layer 6 vigilance grant
// (CR 613.1f). "A Nissa planeswalker" is a planeswalker with the Nissa
// subtype (not a card merely named Nissa), read through the layers.
//
// No simplification.
func init() {
	whileNissa := effects.WhileYouHave(func(g *game.Game, you uuid.UUID) bool {
		return effects.B42ControlsPlaneswalkerNamed(g, you, "Nissa")
	}, effects.SelfCreature)
	vigilance := effects.KeywordGrant(whileNissa, "vigilance")
	vigilance.Label = "Has vigilance as long as you control a Nissa planeswalker."
	effects.Register(effects.Spec{
		OracleID:        "1342bd59-91d3-4a5b-b853-542a52d7da6d",
		Name:            "Guardian of the Great Conduit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			{
				Label:     "Gets +2/+0 as long as you control a Nissa planeswalker.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileNissa,
				Apply:     effects.PumpBy(2, 0),
			},
			vigilance,
		},
	})
}
