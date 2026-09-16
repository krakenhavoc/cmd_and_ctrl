package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Reaper — 3/2 Creature — Zombie Knight for {2}{B}:
//
//	"Whenever a nontoken creature you control dies, this creature
//	deals 1 damage to you and you draw a card."
//
// The nontoken clause is the point: a sacrifice deck full of tokens
// gets nothing from the Reaper, which is exactly the tension the
// card is printed for. It's also the first card to care about the
// token / nontoken split at all — before S21 tokens were
// indistinguishable from cards once they hit the battlefield.
//
// Unlike Blood Artist the Reaper does NOT see its own death: "a
// nontoken creature you control" excludes the source per CR 603.6c
// only when the card says "another", so the Reaper does trigger on
// itself — the damage and draw happen after it's gone.
func init() {
	Register(Spec{
		OracleID:     "e8c7566d-7cc0-48af-a986-83223ec7e06c",
		Name:         "Midnight Reaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller && !IsToken(dead)
			}, "Midnight Reaper — 1 damage to you, draw a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DealDamage{
					Source: ctx.Source(),
					Target: item.Controller,
					Amount: 1,
				}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
