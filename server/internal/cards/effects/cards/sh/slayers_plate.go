package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slayer's Plate — Artifact — Equipment {3} (Shadows over Innistrad, #5543):
//
//	"Equipped creature gets +4/+2.
//	 Whenever equipped creature dies, if it was a Human, create a 1/1 white
//	 Spirit creature token with flying.
//	 Equip {3}"
//
// The dies trigger is read off the creature as it last existed (CR 603.10a):
// it was equipped by this Plate and was a Human then (a changeling counts).
// Both are facts about the past, so the intervening "if" (CR 603.4) holds
// again at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05004b4a-2e2c-4a88-a994-8576684fb64a",
		Name:         "Slayer's Plate",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.EquippedCreatureDied(ev, source) && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Human"))
			}, "Slayer's Plate — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
