package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Charging War Boar — Creature — Boar {1}{R}{G}, 3/1 (Ravnica Allegiance):
//
//	"Haste
//	 As long as you control a Domri planeswalker, this creature gets +1/+1 and
//	 has trample."
//
// Two statics share one condition: the +1/+1 in layer 7c and the trample grant
// in the ability layer. "A Domri planeswalker" is any planeswalker you control
// with the subtype Domri, read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b04b60b-0c4c-4fee-a54b-7052ca5190e1",
		Name:            "Charging War Boar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control a Domri planeswalker, this creature gets +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: chargingWarBoarControlsDomri,
				Apply:     effects.PumpBy(1, 1),
			},
			effects.B16GrantKeywords(chargingWarBoarControlsDomri, "trample"),
		},
	})
}

func chargingWarBoarControlsDomri(target *game.Card, g *game.Game, source *game.Card) bool {
	return effects.SelfCreature(target, g, source) &&
		effects.CountPermanentsYouControl(g, source, func(o *game.Card) bool {
			return o.IsPlaneswalker() && o.HasSubtype("Domri")
		}) > 0
}
