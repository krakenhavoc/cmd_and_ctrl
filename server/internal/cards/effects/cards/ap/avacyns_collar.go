package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avacyn's Collar — Artifact — Equipment {1}:
//
//	"Equipped creature gets +1/+0 and has vigilance.
//	 Whenever equipped creature dies, if it was a Human, create a 1/1 white
//	 Spirit creature token with flying.
//	 Equip {2}"
//
// The dies trigger is read off the creature as it last existed (CR 603.10a):
// it was equipped by the Collar and was a Human then (a changeling counts).
// The intervening "if" (CR 603.4) holds again at resolution.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "842e3835-fff3-4684-bfd1-e7064a1fd2e6",
		Name:         "Avacyn's Collar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.EquippedCreatureDied(ev, source) && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Human"))
			}, "Avacyn's Collar — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
