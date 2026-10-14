package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noosegraf Mob — Creature — Zombie {4}{B}{B}, 0/0:
//
//	"This creature enters with five +1/+1 counters on it.
//	 Whenever a player casts a spell, remove a +1/+1 counter from this
//	 creature. If you do, create a 2/2 black Zombie creature token."
//
// The counters are an entry replacement (b10EntersWithCounters, CR
// 614.1c). The cast trigger is anyone's spell; "if you do" means the
// Zombie comes only when a counter really came off, so a Mob that has
// none left (or has left the battlefield) makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f27fb53f-a983-410a-821b-e48cd8c01f2e",
		Name:         "Noosegraf Mob",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 5, "Noosegraf Mob: enters with five +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnyPlayer, "Noosegraf Mob — remove a +1/+1 counter; if you do, create a 2/2 Zombie",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if ctx.isNewSourceObject(item.SourceCardID) || !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || c.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -1); err != nil {
						return err
					}
					return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
				}),
		},
	})
}
