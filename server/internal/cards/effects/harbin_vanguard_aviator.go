package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harbin, Vanguard Aviator — Legendary Creature — Human Soldier {W}{U}, 3/2
// (The Brothers' War):
//
//	"Flying
//	 Whenever you attack with five or more Soldiers, creatures you control
//	 get +1/+1 and gain flying until end of turn."
//
// Meddling Youths' batch count, restricted to Soldiers: only attack events of
// one declaration by you whose attacker is a Soldier are counted
// (b21CreatureOfSubtypeYouControlAttacked, effective subtypes, so a
// changeling counts), and one trigger fires when there are at least five.
// Harbin need not be among them. The creatures are those you control as it
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d827575f-e364-4a20-887e-3091e5501a4f",
		Name:            "Harbin, Vanguard Aviator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && b21CreatureOfSubtypeYouControlAttacked(ev, source, g, "Soldier")
			}, "Harbin — creatures you control get +1/+1 and gain flying until end of turn (you attacked with five or more Soldiers)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					match := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: match, Power: 1, Toughness: 1, Label: "Harbin — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"flying"}, Label: "Harbin — flying"}.Apply(ctx)
				}),
				ThatManyEventsAtLeast(5)),
		},
	})
}
