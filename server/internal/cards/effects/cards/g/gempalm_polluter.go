package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	cycled := effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Gempalm Polluter — target player loses life for each Zombie", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.HasSubtype(ctx.TextSubtype("Zombie")) {
				n++
			}
		}
		return (effects.TargetPlayerLosesLife{N: n}).Apply(ctx)
	}))
	cycled.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{OracleID: "bd532f8d-789f-4ee1-a2f3-ae30e8282c76", Name: "Gempalm Polluter", Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.Cycling("{B}{B}")},
		Triggered: []game.TriggeredAbility{effects.Optional(cycled, "Gempalm Polluter — have that player lose life?")},
	})
}
