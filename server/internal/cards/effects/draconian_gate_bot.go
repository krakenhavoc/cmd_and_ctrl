package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	bot := WhenThisEnters("Draconian Gate-Bot — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bot.Modes = ChooseOne(
		ModeDoing("Open an Attraction.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return OpenAttractions{N: 1}.Apply(ctx)
		}),
		ModeDoing("Destroy target Attraction.", TargetPermanent("target Attraction", Attraction()),
			DestroyTheModesTarget),
	)
	Register(Spec{
		OracleID:     "f73ba6ae-87c1-4e5d-ae52-59d11d7d6ba6",
		Name:         "Draconian Gate-Bot",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{bot},
	})
}
