package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Man, Master of Machines — {2}{U}{R} Legendary Artifact Creature —
// Human Hero 1/4 (Marvel Super Heroes, #4455):
//
//	"Flying, vigilance
//	 Iron Man gets +1/+0 for each other artifact you control.
//	 Whenever Iron Man attacks, if an artifact entered the battlefield
//	 under your control this turn, draw a card."
//
// The bonus is a layer 7c modification read off the board every time the
// layers are recomputed, counting artifacts by their post-layer types and
// leaving Iron Man himself out. The attack trigger's "if" is intervening
// (CR 603.4): it is checked as the attack is declared and again as the
// trigger resolves, against the turn tally of artifacts that entered under
// your control (CR 400.7, tokens included, Iron Man himself counted if he
// came in this turn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94bc796d-8c04-4835-8a99-7e5dce9540ee",
		Name:            "Iron Man, Master of Machines",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Iron Man gets +1/+0 for each other artifact you control.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += effects.CountPermanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsArtifact()
				})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "an artifact entered the battlefield under your control this turn",
				Test: func(ctx *effects.Context) bool {
					return ctx.Game.EnteredWithCardTypeThisTurn(ctx.Controller(), "artifact") > 0
				},
			}, effects.On(game.EventAttack, effects.AllOf(effects.ThisAttacked, mshA4ArtifactEnteredUnderYou),
				"Iron Man — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}

// mshA4ArtifactEnteredUnderYou is the trigger-time half of the intervening
// "if an artifact entered the battlefield under your control this turn".
func mshA4ArtifactEnteredUnderYou(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
}
