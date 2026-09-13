package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Familiar — 2/2 Artifact Creature — Fox for {3}:
//
//	"When Filigree Familiar enters the battlefield, you gain 2 life.
//	When Filigree Familiar dies, draw a card.
//	{2}, Sacrifice Filigree Familiar: Add one mana of any color."
//
// S19 sub-PR 4 shipped only the dies half ("draw a card"). The
// #338 stale-simplification sweep adds the ETB-lifegain half, which
// had been "deferred to a later batch" since before ETB triggers
// were routine — Archivist of Oghma and Authority of the Consuls
// were already doing exactly this shape.
//
// cardDied gates the dies trigger to graveyard-only: a bounced or
// exiled Familiar does not draw. The draw happens when the trigger
// resolves.
//
// Declared simplification, still real: "{2}, Sacrifice Filigree
// Familiar: Add one mana of any color" is NOT implemented. A mana
// ability's cost has no mana component — ManaAbilityCost is
// tap / sacrifice / sacrifice-another / life — so the {2} cannot be
// expressed. (Lotus Petal works only because its cost is tap+sac
// with no mana in it.) Routing this through the CR 602 activated
// path instead would be wrong: mana abilities don't use the stack.
func init() {
	Register(Spec{
		OracleID:     "b544f690-e4bf-4a5b-984d-9256518fd574",
		Name:         "Filigree Familiar",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{2}, Sacrifice: Add one mana of any color\" ability is missing — the Fox can't be cracked for mana."},
		Triggered: []game.TriggeredAbility{{
			// "When this enters the battlefield, you gain 2 life."
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Filigree Familiar — gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
					})
			},
		}, {
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Filigree Familiar — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
