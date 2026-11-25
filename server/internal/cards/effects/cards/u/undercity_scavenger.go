package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Scavenger — Creature — Ogre Warrior {3}{B}, 3/3 (Ravnica
// Allegiance, #5521):
//
//	"When this creature enters, you may sacrifice another creature. If you
//	 do, put two +1/+1 counters on this creature, then scry 2."
//
// The sacrifice is asked as the ability resolves, over whatever creatures
// you control then; nothing to sacrifice asks nothing. The counters go on
// only while the Scavenger is still the same object on the battlefield; the
// scry happens either way once a creature was sacrificed.
//
// No simplification.
func init() {
	const name = "Undercity Scavenger"
	purpose := game.Purpose{}
	purpose.Scry = 2
	effects.Register(effects.Spec{
		OracleID:     "5146ae31-9578-46e3-be78-eb90e4ba48cd",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — you may sacrifice another creature; if you do, two +1/+1 counters on this creature, then scry 2",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, name+" — sacrifice another creature to put two +1/+1 counters on this creature and scry 2?",
						effects.Creature(), true, undercityScavengerSacrificed)
				}),
		},
	})
}

var undercityScavengerSacrificed = effects.RegisterSacrificedOne("prompt-run/undercity-scavenger-counters-and-scry", func(ctx *effects.Context, _ uuid.UUID) error {
	if !effects.SourceIsNewObject(ctx.Game, ctx.Item) && effects.OnBattlefield(ctx.Game, ctx.Source()) {
		if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
