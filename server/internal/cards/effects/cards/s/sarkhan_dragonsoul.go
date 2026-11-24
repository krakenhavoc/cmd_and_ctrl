package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan, Dragonsoul — Legendary Planeswalker — Sarkhan {4}{R}{R}, loyalty 5:
//
//	"+2: Sarkhan deals 1 damage to each opponent and each creature your
//	 opponents control.
//	 −3: Sarkhan deals 4 damage to target player or planeswalker.
//	 −9: Search your library for any number of Dragon creature cards, put them
//	 onto the battlefield, then shuffle."
//
// +2 hits only your opponents and their creatures. −9 is a search with no
// ceiling, and the cards enter together; the shuffle follows even if none was
// found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2c1e44c-2a19-466f-a6d5-4cb71a940b86",
		Name:            "Sarkhan, Dragonsoul",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Sarkhan deals 1 damage to each opponent and each creature your opponents control.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachOpponentAndEachCreatureTheyControl(effects.NewContext(g, item), 1)
				},
			},
			{
				Label:   "−3: Sarkhan deals 4 damage to target player or planeswalker.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Effect:  effects.DealDamageToTheTarget(4),
			},
			{
				Label: "−9: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						// "Dragon" is a word of the source's text (CR 612.2).
						Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Dragon") },
						Dest:      game.ZoneBattlefield,
						Unbounded: true,
						Shuffle:   true,
						Reason:    "Sarkhan, Dragonsoul — any number of Dragon creature cards onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
