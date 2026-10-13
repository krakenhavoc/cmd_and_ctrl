package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kangee, Aerie Keeper — Legendary Creature — Bird Wizard {2}{W}{U}, 2/2
// (#3232):
//
//	"Kicker {X}{2} (You may pay an additional {X}{2} as you cast this
//	 spell.)
//	 Flying
//	 When Kangee enters, if it was kicked, put X feather counters on it.
//	 Other Bird creatures get +1/+1 for each feather counter on Kangee."
//
// The kicker is {X}{2}: the {2} is a flat part of the price and X is the
// caster's one announced number, with no floor (a kicked Kangee at X = 0
// is legal and puts no counters). The counters go on after Kangee has
// landed, from a trigger the table can respond to, reading the X
// announced for the spell (CR 107.3m). The lord buffs every OTHER Bird
// at the table, not only yours, because the card has no controller
// clause.
//
// No simplification.
func init() {
	const label = "Kangee, Aerie Keeper — put X feather counters on it"
	otherBirds := TribeFilter{Tribes: []string{"Bird"}, Others: true}
	Register(Spec{
		OracleID:        "555f9e75-8892-4c8e-9c5e-8ab14a589118",
		Name:            "Kangee, Aerie Keeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{KickerX("{X}{2}", 0)},
		Static: []game.StaticAbility{
			TribalScalingAnthem(otherBirds, func(source *game.Card, _ *game.Game) int {
				return source.Counters["feather"]
			}),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: AllOf(Self, ThisKickedAtLeast(1)),
			Key:       label,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "feather", N: x}.Apply(NewContext(g, item))
			},
		}},
	})
}
