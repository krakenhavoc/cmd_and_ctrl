package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draconian Gate-Bot — Artifact Creature — Robot {5}:
//
//	"When this creature enters, choose one —
//	 • Open an Attraction. (Put the top card of your Attraction deck
//	   onto the battlefield.)
//	 • Destroy target Attraction. (It's put into its owner's junkyard.)"
//
// A modal entry trigger (CR 603.3c). The second bullet's Attraction goes
// to its owner's junkyard because the rules send it there (CR 717.6, the
// engine's built-in replacement), not because this card says so: the
// parenthetical is reminder text. With no Attraction on the battlefield
// the second bullet is not on offer.
//
// No simplification.
func init() {
	bot := effects.WhenThisEnters("Draconian Gate-Bot — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bot.Modes = effects.ChooseOne(
		effects.ModeDoing("Open an Attraction.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.OpenAttractions{N: 1}.Apply(ctx)
		}),
		effects.ModeDoing("Destroy target Attraction.", effects.TargetPermanent("target Attraction", effects.Attraction()),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "f73ba6ae-87c1-4e5d-ae52-59d11d7d6ba6",
		Name:         "Draconian Gate-Bot",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{bot},
	})
}
