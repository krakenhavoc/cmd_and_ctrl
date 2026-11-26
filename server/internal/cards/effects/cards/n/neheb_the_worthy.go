package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neheb, the Worthy — {1}{B}{R} Legendary Creature — Minotaur Warrior 2/2
// (Amonkhet, S82 old batch 248):
//
//	"First strike
//	 Other Minotaurs you control have first strike.
//	 As long as you have one or fewer cards in hand, Minotaurs you control get
//	 +2/+0.
//	 Whenever Neheb deals combat damage to a player, each player discards a
//	 card."
//
// The grant (CR 613.1f) is to the controller's other Minotaurs; the pump is a
// layer 7c modification (CR 613.4c) of every Minotaur they control, Neheb
// included, recomputed on every change to the hand (DependsOnHandSize). The
// discard is each player in turn order, the controller among them.
//
// No simplification.
func init() {
	minotaur := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Minotaur"))
	}
	otherMinotaur := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return minotaur(target, g, source) && target.InstanceID != source.InstanceID
	}
	effects.Register(effects.Spec{
		OracleID:        "092847a7-a238-4acf-9dcd-be6f3e6a6b93",
		Name:            "Neheb, the Worthy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(otherMinotaur, "first strike"),
			{
				Label:             "As long as you have one or fewer cards in hand, Minotaurs you control get +2/+0.",
				Layer:             game.Layer7PT,
				SubLayer:          game.SubLayer7C_Modify,
				DependsOnHandSize: true,
				AppliesTo:         effects.WhileYouHave(effects.YouHaveAtMostCardsInHand(1), minotaur),
				Apply:             effects.PumpBy(2, 0),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Neheb, the Worthy — each player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
						Source: item.SourceCardID, N: 1, Question: "Neheb, the Worthy — discard a card",
					})
					return nil
				}),
		},
	})
}
