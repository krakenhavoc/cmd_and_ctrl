package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Korvold, Fae-Cursed King — Legendary Creature — Dragon Noble
// {2}{B}{R}{G}, 4/4:
//
//	"Flying
//	 Whenever Korvold enters or attacks, sacrifice another permanent.
//	 Whenever you sacrifice a permanent, put a +1/+1 counter on
//	 Korvold and draw a card."
//
// The commander this sprint's theme deck is built around, and the
// card that needs every piece of S21 at once: a mandatory sacrifice
// with a choice, an EventSacrifice payoff, a counter and a draw. The
// two abilities feed each other — the first one pays for the second,
// every combat, without any other card on the board.
//
// Three details worth naming:
//
//   - "Sacrifice ANOTHER permanent" is a choice the CONTROLLER makes,
//     so it queues a sacrifice picker rather than the engine picking.
//     Korvold excludes himself via NotSelf; with no other permanent
//     the prompt is skipped and nothing is sacrificed (CR 701.17b).
//   - It is MANDATORY. A Korvold attacking into an empty board eats
//     a land. There is no "you may" on the printed card, and adding
//     one would make it strictly better than printed.
//   - The second ability sees the first one's sacrifice: the picker's
//     answer emits EventSacrifice, which triggers the counter and the
//     draw. It also sees every OTHER sacrifice you make — a Treasure
//     cracked for mana, a creature fed to Goblin Bombardment, a
//     Blood token discarded away.
//
// "Enters or attacks" is one ability with two conditions (Sun Titan's
// shape) rather than two declarations.
func init() {
	Register(Spec{
		OracleID:        "9ae669dd-7e60-4649-b96e-35da28be641a",
		Name:            "Korvold, Fae-Cursed King",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					self := source.InstanceID
					return game.NewTriggeredItem(source, "Korvold — sacrifice another permanent",
						func(g *game.Game, item *game.StackItem) error {
							g.PlayerSacrificesForEffect(
								item.SourceCardID,
								item.Controller,
								sacrificeSpec("another permanent", NotSelf(self)),
								"Korvold — sacrifice another permanent",
							)
							return nil
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Korvold — +1/+1 counter and draw a card",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							if err := (AddCounter{
								Target: ctx.Source(),
								Kind:   game.CounterPlusOne,
								N:      1,
							}).Apply(ctx); err != nil {
								return err
							}
							return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
						})
				},
			},
		},
	})
}
