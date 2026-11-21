package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beanstalk Wurm // Plant Beans — Creature — Plant Wurm {4}{G}, 5/4 //
// Sorcery — Adventure {1}{G} (Wilds of Eldraine, #4471):
//
//	Beanstalk Wurm — "Reach"
//	Plant Beans — "You may play an additional land this turn. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"), as Ferocious Werefox's do; the Adventure lifecycle (exile on
// resolution, cast the creature later) is the engine's. Plant Beans is
// Explore's one-shot land-drop grant without the draw, so it expires at
// the turn boundary and two grants stack.
//
// No simplification.
const BeanstalkWurmOracleID = "a2192c80-c48f-44e2-807c-2f9ee1aa1152"

func init() {
	effects.Register(effects.Spec{
		OracleID:        BeanstalkWurmOracleID,
		Name:            "Beanstalk Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
	effects.Register(effects.Spec{
		OracleID:     BeanstalkWurmOracleID + "#1",
		Name:         "Plant Beans",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}
