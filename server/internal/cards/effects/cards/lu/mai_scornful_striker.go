package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mai, Scornful Striker — Legendary Creature — Human Noble Ally {1}{B},
// 2/2:
//
//	"First strike
//	 Whenever a player casts a noncreature spell, they lose 2 life."
//
// First strike rides PrintedKeywords. The trigger is
// castersLoseLife (caster_loses_life.go) over every player, Mai's
// controller included. "Noncreature" is read off the spell on the
// stack, so an artifact creature spell is a creature spell and a Saga,
// an Equipment or a sorcery all trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "953a2bd3-5bca-41fc-8785-66b2d7fa381a",
		Name:            "Mai, Scornful Striker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.CastersLoseLife("Mai, Scornful Striker — that player loses 2 life", 2,
				func(spell game.Card) bool { return !spell.IsCreature() }),
		},
	})
}
