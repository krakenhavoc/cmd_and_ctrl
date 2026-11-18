package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Juniper Order Ranger — Creature — Human Knight Ranger {3}{G}{W}, 2/4
// (Magic Commander Collection, #5908):
//
//	"Whenever another creature you control enters, put a +1/+1 counter on
//	 that creature and a +1/+1 counter on this creature."
//
// One trigger, two counters. The entering creature is the event's card; each
// half is skipped if its object has left the battlefield by resolution (the
// Ranger's half also if it is a new object, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18ca77c9-e7d2-4afc-a72b-98772578fad7",
		Name:         "Juniper Order Ranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Juniper Order Ranger — put a +1/+1 counter on that creature and a +1/+1 counter on this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if entering := item.Trigger.Event.CardID; onBattlefield(g, entering) {
						if err := (AddCounter{Target: entering, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
