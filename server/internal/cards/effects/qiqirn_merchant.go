package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Qiqirn Merchant — Creature — Beast Citizen {2}{U}, 1/4 (Final
// Fantasy, #4462):
//
//	"{1}, {T}: Draw a card, then discard a card.
//	 {7}, {T}, Sacrifice this creature: Draw three cards. This ability
//	 costs {1} less to activate for each Town you control."
//
// Captain of Umbar's loot, and Balamb Garden's per-Town reduction
// (CostsLessEach) over every Town its controller controls: the Merchant
// is no Town, so none is "other".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "560cb970-1c1b-4ba7-8352-bc67b4ea7843",
		Name:         "Qiqirn Merchant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
			},
			{
				Label:   "{7}, {T}, Sacrifice this creature: Draw three cards. This ability costs {1} less to activate for each Town you control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{7}"), TapCost(), SacrificeThis()),
				CostModifiers: []game.CostModifier{
					CostsLessEach(qiqirnTowns, "This ability costs {1} less to activate for each Town you control."),
				},
				Effect: Do(DrawCards{N: 3}),
			},
		},
	})
}

// qiqirnTowns counts the Towns the ability's controller controls.
func qiqirnTowns(q game.CostQuery) int {
	town := MatchLandSubtype("Town")
	return countBattlefield(q, func(c game.Card) bool { return c.Controller == q.Controller && town(c) })
}
