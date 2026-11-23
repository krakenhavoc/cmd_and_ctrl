package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Capitoline Triad — Legendary Creature — God Artificer {10}, 7/7
// (Assassin's Creed, #5948):
//
//	"Those Who Came Before — This spell costs {1} less to cast for each
//	 historic card in your graveyard. (Artifacts, legendaries, and Sagas are
//	 historic.)
//	 Exile any number of historic cards from your graveyard with total mana
//	 value 30 or greater: You get an emblem with "Creatures you control have
//	 base power and toughness 9/9.""
//
// The reduction is a self cost modifier counting the historic cards in the
// caster's graveyard as the spell is priced (CR 601.2f); the Triad itself is
// in the hand and never counts. The ability's cost is an exile of any number
// of historic cards from the graveyard whose mana values total 30 or more,
// named by the activator at announce (CR 602.2b); the exile is a cost, so the
// ability can be activated at instant speed with no other cost. Its effect is
// the emblem, whose layer 7b static sets base power and toughness of every
// creature its owner controls to 9/9 (CR 613.4b); counters, pumps and other
// non-base effects still apply on top, and the emblem never goes away
// (CR 114.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0321049-a3ed-4378-b607-7fb190b6f0fd",
		Name:         "The Capitoline Triad",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.GapB36HistoricCardsInYourGraveyard,
				"Those Who Came Before — this spell costs {1} less to cast for each historic card in your graveyard."),
		},
		Emblem: &effects.EmblemSpec{
			Label: "The Capitoline Triad emblem",
			Text:  "Creatures you control have base power and toughness 9/9.",
			Static: []game.StaticAbility{{
				Label:     "Creatures you control have base power and toughness 9/9.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: effects.CreaturesTheEmblemsOwnerControls,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 9, 9
				},
			}},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile any number of historic cards from your graveyard with total mana value 30 or greater: You get an emblem with \"Creatures you control have base power and toughness 9/9.\"",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.GapB36ExileAnyNumberWithTotalManaValue("historic cards", 30, effects.B09IsHistoric),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
