package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidecaller Mentor — Creature — Rat Wizard {1}{U}{B}, 3/3 (Bloomburrow,
// #4467):
//
//	"Menace
//	 Threshold — When this creature enters, if there are seven or more
//	 cards in your graveyard, return up to one target nonland permanent
//	 to its owner's hand."
//
// Crypt Feaster's intervening "if" (CR 603.4) on Dundoolin Weaver's
// targeted entry trigger: the seven cards are counted as the Mentor
// enters and again as the trigger resolves. The still-legal pick returns
// to its owner's hand (bounceTheTargetToHand); choosing none returns
// nothing. Menace rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e75e40a5-a9ae-4789-96c1-0e19d1ce59c5",
		Name:            "Tidecaller Mentor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.CondBoard{
				Name: "there are seven or more cards in your graveyard",
				Test: func(ctx *effects.Context) bool { return effects.FraThresholdMet(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.FraThresholdMet(g, source.Controller)
			}, "Tidecaller Mentor — threshold: return up to one target nonland permanent to its owner's hand", effects.BounceTheTargetToHand)),
				effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1)),
		},
	})
}
