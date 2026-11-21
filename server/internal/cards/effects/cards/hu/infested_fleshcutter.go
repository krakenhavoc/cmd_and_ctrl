package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infested Fleshcutter — Artifact — Equipment {1}{W}:
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature attacks, create a 1/1 colorless Phyrexian Mite
//	 artifact creature token with toxic 1 and 'This token can't block.'
//	 Equip {2}{W}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9c03ed6-34dd-46f7-ab41-8004348328cc",
		Name:         "Infested Fleshcutter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Infested Fleshcutter — create a Phyrexian Mite", effects.Do(effects.CreateToken{Template: effects.PhyrexianMiteToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}{W}")},
	})
}
