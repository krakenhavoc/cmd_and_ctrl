package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surrak and Goreclaw — Legendary Creature — Human Bear {4}{G}{G}, 6/5:
//
//	"Trample
//	 Other creatures you control have trample.
//	 Whenever another nontoken creature you control enters, put a +1/+1
//	 counter on it. It gains haste until end of turn."
//
// Aggressive Mammoth's trample grant to every other creature the
// controller has, and Rocksteady's "another nontoken creature you control
// enters" condition. The counter and the haste are aimed at the creature
// named by the trigger's event; one that has already left does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "95be5f72-1f81-4ec6-9bd9-f0b07bfc874b",
		Name:            "Surrak and Goreclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Surrak and Goreclaw — put a +1/+1 counter on it; it gains haste until end of turn", surrakEntered),
		},
	})
}

func surrakEntered(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	id := item.Trigger.Event.CardID
	if !onBattlefield(g, id) {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"haste"},
		Label:    "Surrak and Goreclaw — haste",
	}.Apply(ctx)
}
