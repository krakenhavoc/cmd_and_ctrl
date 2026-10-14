package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shelinda, Yevon Acolyte — Legendary Creature — Human Cleric {G}{W},
// 2/2:
//
//	"Lifelink
//	 Whenever another creature you control enters, put a +1/+1 counter
//	 on that creature if its power is less than Shelinda's power.
//	 Otherwise, put a +1/+1 counter on Shelinda."
//
// Both powers are read as the trigger resolves (CR 608.2h): each
// permanent as it is now while it is still on the battlefield, and as it
// last existed once it has gone. A counter is only ever put on a
// permanent that is still the same object on the battlefield, so a
// creature that left in response gets nothing (and the "otherwise"
// branch does not move the counter to Shelinda), and a Shelinda that
// left gets nothing either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b06c9a4b-f309-45b3-ae5b-0ce3f067a5cb",
		Name:            "Shelinda, Yevon Acolyte",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Shelinda, Yevon Acolyte — a +1/+1 counter on the smaller of the two",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					entered, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					shelinda, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					if entered.Power < shelinda.Power {
						if entered.Left {
							return nil
						}
						return g.AddCounterByForEffect(item.Controller, item.Trigger.Event.CardID, game.CounterPlusOne, 1)
					}
					if shelinda.Left {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
