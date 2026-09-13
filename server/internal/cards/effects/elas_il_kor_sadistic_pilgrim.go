package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elas il-Kor, Sadistic Pilgrim — Legendary Creature — Phyrexian Kor
// Cleric {W}{B}, 2/2 (EDHREC rank 615):
//
//	"Deathtouch
//	 Whenever another creature you control enters, you gain 1 life.
//	 Whenever another creature you control dies, each opponent loses
//	 1 life."
//
// Soul Warden and Zulaport Cutthroat on one body. Two triggers, both
// "ANOTHER creature you control" — Elas's own entry and own death are
// excluded, unlike Zulaport's "this creature or another". The drain
// is life loss, not damage, so no prevention applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4d33e21-d3fe-4d5b-903c-b75e859d6f7f",
		Name:            "Elas il-Kor, Sadistic Pilgrim",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature()
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Elas il-Kor — you gain 1 life",
						func(g *game.Game, item *game.StackItem) error {
							return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false // "another"
					}
					dead, ok := diedCreature(ev, g)
					return ok && dead.Controller == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Elas il-Kor — each opponent loses 1 life",
						func(g *game.Game, item *game.StackItem) error {
							return eachOpponentLosesLife(g, item, 1)
						})
				},
			},
		},
	})
}
