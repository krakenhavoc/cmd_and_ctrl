package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krenko, Tin Street Kingpin — 1/2 Legendary Creature — Goblin for
// {2}{R}:
//
//	"Whenever Krenko attacks, put a +1/+1 counter on it, then create
//	a number of 1/1 red Goblin creature tokens equal to Krenko's
//	power."
//
// S22 attack triggers: the "whenever THIS creature attacks" shape,
// and the one where the printed word "then" is load-bearing. The
// counter goes on first and the token count is read AFTER it, so a
// freshly-cast 1/2 Krenko makes two Goblins on its first swing, not
// one. Reading power before the counter is the classic mis-read.
//
// Power is read live off the battlefield at resolution rather than
// captured in Build, because everything between declaration and
// resolution counts — a pump spell, an anthem, a second Krenko
// trigger from an earlier combat phase.
//
// Sandbox simplification: if Krenko has left the battlefield by the
// time the trigger resolves, the ability makes no tokens. Paper uses
// last-known information (CR 608.2h) and would still make Goblins
// equal to its last power; the LKI snapshot the engine keeps is
// scoped to the dying card's own LTB triggers and isn't reachable
// from a resolution callback.
func init() {
	Register(Spec{
		OracleID: "e8065e1d-e937-4b56-8011-78f0d07328a0",
		Name:     "Krenko, Tin Street Kingpin",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Krenko, Tin Street Kingpin — +1/+1 counter, then Goblins",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (AddCounter{
							Target: item.SourceCardID,
							Kind:   game.CounterPlusOne,
							N:      1,
						}).Apply(ctx); err != nil {
							return err
						}
						krenko, ok := g.LookupCardForEffect(item.SourceCardID)
						if !ok {
							return nil
						}
						return CreateToken{
							Controller: item.Controller,
							Template:   RedGoblinToken(),
							N:          krenko.CurrentPower(),
						}.Apply(ctx)
					})
			},
		}},
	})
}
