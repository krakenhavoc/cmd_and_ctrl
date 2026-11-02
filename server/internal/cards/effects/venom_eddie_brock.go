package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venom, Eddie Brock — Legendary Creature — Symbiote Villain {4}{B}{B},
// 6/4 (Marvel's Spider-Man, #4460):
//
//	"Menace
//	 Whenever another creature dies, put a +1/+1 counter on Venom. If
//	 that creature was a Villain, draw a card."
//
// One trigger. Whether the dead creature was a Villain is read as it last
// existed on the battlefield (CR 603.10a, Event.LastKnownSubtypes) when
// the trigger is built, and carried on the item. The counter lands only
// if Venom is still on the battlefield; the draw does not need him
// (CR 608.2b, 608.2h).
func init() {
	const label = "Venom, Eddie Brock — +1/+1 counter on Venom; draw a card if it was a Villain"
	Register(Spec{
		OracleID:        "9132d7e6-7190-45e1-9c57-8fe2e5d0f2f2",
		Name:            "Venom, Eddie Brock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventLTB},
			AppliesTo: AnotherCreatureDied,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				if dead, ok := g.LookupCardForEffect(ev.CardID); ok && leftAsSubtype(ev, dead, source.TextSubtype("Villain")) {
					item.Params.Amount = 1
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if item.Params.Amount == 0 {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			},
		}},
	})
}
