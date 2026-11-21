package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mantle of Tides — Artifact — Equipment {U} (Dominaria Remastered):
//
//	"Equipped creature gets +1/+2.
//	 Whenever you draw your second card each turn, attach this Equipment
//	 to target creature you control.
//	 Equip {3}"
//
// Alandra's second-draw tally with Bramble Armor's attach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56a80bd3-b81b-428d-96e1-f03ab6199cf3",
		Name:         "Mantle of Tides",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Mantle of Tides — attach this Equipment to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
