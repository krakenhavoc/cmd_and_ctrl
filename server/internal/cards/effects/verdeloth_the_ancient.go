package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdeloth the Ancient — Legendary Creature — Treefolk {4}{G}{G}, 4/7
// (#3232):
//
//	"Kicker {X} (You may pay an additional {X} as you cast this spell.)
//	 Saproling creatures and other Treefolk creatures get +1/+1.
//	 When Verdeloth enters, if it was kicked, create X 1/1 green
//	 Saproling creature tokens."
//
// The lord has no controller clause: it buffs EVERY Saproling and every
// other Treefolk at the table, which is the printed drawback
// (TribeFilter without YoursOnly). Verdeloth is a Treefolk, so "other"
// excludes it. A kicked Verdeloth with X = 0 is legal (the card prints
// no floor) and makes no tokens. The enters trigger is an intervening if
// (CR 603.4) reading the X announced for the spell (CR 107.3m) as it
// goes on the stack, so the tokens are made by a trigger the table can
// respond to.
//
// No simplification.
func init() {
	saprolings := TribeFilter{Tribes: []string{"Saproling"}}
	otherTreefolk := TribeFilter{Tribes: []string{"Treefolk"}, Others: true}
	const label = "Verdeloth the Ancient — create X 1/1 green Saproling tokens"
	Register(Spec{
		OracleID:      "8222321f-b74d-4f74-9ee7-79455433345a",
		Name:          "Verdeloth the Ancient",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{KickerX("{X}", 0)},
		Static: []game.StaticAbility{
			TribalAnthem(saprolings, 1, 1),
			TribalAnthem(otherTreefolk, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: AllOf(Self, ThisKickedAtLeast(1)),
			Key:       label,
			// A fill-in Build (ADR 0041 P9): X is a fact of the moment
			// the trigger fired, stamped as a plain int rather than a
			// closure over the card.
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
				return CreateToken{Controller: item.Controller, Template: b11GreenSaprolingToken(), N: x}.Apply(NewContext(g, item))
			},
		}},
	})
}
