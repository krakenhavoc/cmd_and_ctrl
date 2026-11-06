package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const beanstalkWurmOracleID = "a2192c80-c48f-44e2-807c-2f9ee1aa1152"

func init() {
	Register(Spec{
		OracleID:        beanstalkWurmOracleID,
		Name:            "Beanstalk Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
	Register(Spec{
		OracleID:     beanstalkWurmOracleID + "#1",
		Name:         "Plant Beans",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}
