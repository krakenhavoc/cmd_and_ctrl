package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

func init() {
	cycled := InGraveyard(On(game.EventCycle, Self, "Gempalm Polluter — target player loses life for each Zombie", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.HasSubtype(ctx.TextSubtype("Zombie")) {
				n++
			}
		}
		return (TargetPlayerLosesLife{N: n}).Apply(ctx)
	}))
	cycled.Targets = TargetPlayer("target player")
	Register(Spec{OracleID: "bd532f8d-789f-4ee1-a2f3-ae30e8282c76", Name: "Gempalm Polluter", Completeness: CompletenessFull,
		Activated: []ActivatedAbility{Cycling("{B}{B}")},
		Triggered: []game.TriggeredAbility{Optional(cycled, "Gempalm Polluter — have that player lose life?")},
	})
}
