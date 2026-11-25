package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soaring Thought-Thief — Creature — Human Rogue {U}{B}, 1/3 (Zendikar
// Rising, #5508):
//
//	"Flash
//	 Flying
//	 As long as an opponent has eight or more cards in their graveyard,
//	 Rogues you control get +1/+0.
//	 Whenever one or more Rogues you control attack, each opponent mills two
//	 cards."
//
// The pump is a layer 7c static over the Rogues its controller controls
// (itself included), live on any opponent's graveyard count (Expedition
// Lookout's read). The attack trigger is batched (CR 603.2c): one trigger per
// declaration however many Rogues attack, and each opponent mills two.
//
// No simplification.
func init() {
	const label = "Soaring Thought-Thief — each opponent mills two cards"
	effects.Register(effects.Spec{
		OracleID:        "5ed445c1-8a04-4fab-91ef-4d52db95c3b5",
		Name:            "Soaring Thought-Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static: []game.StaticAbility{{
			Label:    "As long as an opponent has eight or more cards in their graveyard, Rogues you control get +1/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Rogue")) &&
					effects.S82OpponentHasGraveyardCards(g, source.Controller, 8)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power++ },
		}},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.AttackerHasSubtype(g, ev, source.TextSubtype("Rogue"))
			}, label, func(g *game.Game, item *game.StackItem) error { return effects.B12EachOpponentMills(g, item, 2) })),
		},
	})
}
