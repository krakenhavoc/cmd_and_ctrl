package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enemy of Enlightenment — Enchantment Creature — Demon {5}{B}, 5/5 (Theros
// Beyond Death, S82 OLD, #5514):
//
//	"Flying
//	 This creature gets -1/-1 for each card in your opponents' hands.
//	 At the beginning of your upkeep, each player discards a card."
//
// The modification is a layer 7 static of the source alone, counted live over
// every opponent's hand (CR 613.4c) with the hand-size invalidation hint
// (Demon's Jester's), so a draw or discard re-reads it. It can take the
// Demon to 0 toughness, which the state-based actions then deal with
// (CR 704.5f). The upkeep trigger is Cunning Lethemancer's: each player in
// turn order discards a card of their choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "783f059f-90d0-4148-898b-9e694b22ac0a",
		Name:            "Enemy of Enlightenment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:             "This creature gets -1/-1 for each card in your opponents' hands.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			DependsOnHandSize: true,
			AppliesTo:         effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.Seats {
					if p != nil && p.ID != source.Controller && !p.Eliminated && p.Hand != nil {
						n += p.Hand.Size()
					}
				}
				c.Power -= n
				c.Toughness -= n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Enemy of Enlightenment — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Enemy of Enlightenment — discard a card",
				})
				return nil
			}),
		},
	})
}
