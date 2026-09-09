package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zulaport Cutthroat — 1/1 Creature — Human Rogue Ally for {1}{B}:
//
//	"Whenever this creature or another creature you control dies,
//	each opponent loses 1 life and you gain 1 life."
//
// Blood Artist's untargeted cousin, narrower in what triggers it
// (your creatures only) and wider in what it hits (every opponent).
// You gain exactly 1 regardless of the table size — the life gain
// isn't per-opponent.
func init() {
	Register(Spec{
		OracleID: "76b003e0-15af-4f22-bdf2-1ade5430964a",
		Name:     "Zulaport Cutthroat",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Zulaport Cutthroat — each opponent loses 1",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, opp := range ctx.Opponents() {
							if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -1); err != nil {
								return err
							}
						}
						return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
					})
			},
		}},
	})
}
