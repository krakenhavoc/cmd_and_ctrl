package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivy Lane Denizen — Creature — Elf Warrior {3}{G}, 2/3 (EDHREC rank
// 2462):
//
//	"Whenever another green creature you control enters, put a +1/+1
//	 counter on target creature."
//
// The Elf deck's counter engine. One targeted trigger per green
// creature that enters under the controller's control — the Denizen
// itself excluded, a token included, and "green" read off the
// creature's effective colours — with "target creature" the S20
// clause: any creature on the battlefield, the entering one and the
// Denizen itself both legal picks, chosen when the trigger goes on
// the stack. The counter goes on through AddCounter, so a doubler
// applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5da5ad6-4ed2-4041-a983-76a8c87fa109",
		Name:         "Ivy Lane Denizen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23AnotherGreenCreatureYouControlEntered(ev, source, g)
			},
			Targets: effects.TargetCreature("target creature"),
			Key:     "Ivy Lane Denizen — put a +1/+1 counter on target creature",
			Effect:  effects.PlusOneCounterOnChosenTargets,
		}},
	})
}
