package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grakk the Pacifist — Legendary Creature — Orc {1}{W}, 4/3 (Mystery Booster
// Commander, #5948):
//
//	"Grakk can't attack or block.
//	 Whenever you cast an Aura spell that targets a creature an opponent
//	 controls, draw a card. This ability triggers only once each turn."
//
// The first sentence is a layer-6 restriction on itself (RestrictSelf). The
// trigger reads the targets the Aura announced (CR 601.2c), as Pearl-Ear does:
// a creature on the battlefield under an opponent's control. The "only once
// each turn" is TriggersOnlyOnceEachTurn, spent when it triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56277f60-67b0-41fa-a363-c64e2d794d01",
		Name:         "Grakk the Pacifist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantAttackOrBlock)},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.On(game.EventCast, effects.GapB05YouCastAnAuraOnAnOpponentsCreature,
				"Grakk the Pacifist — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
