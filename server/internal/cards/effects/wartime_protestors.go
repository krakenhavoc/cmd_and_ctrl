package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wartime Protestors — Creature — Human Rebel Ally {3}{R}, 4/4:
//
//	"Haste
//	 Whenever another Ally you control enters, put a +1/+1 counter on
//	 that creature and it gains haste until end of turn."
//
// The trigger is Avatar Enthusiasts' (another Ally entering under your
// control); its body acts on the creature that entered, and only while
// that is still the same object on the battlefield (CR 400.7): an Ally
// that left and came back in response is a new object and gets neither.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "6557813b-4ee7-4881-a37c-10c8ea097360",
		Name:            "Wartime Protestors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Wartime Protestors — put a +1/+1 counter on that creature and it gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					id := item.Trigger.Event.CardID
					if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"haste"},
						Label:    "Wartime Protestors — haste until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
