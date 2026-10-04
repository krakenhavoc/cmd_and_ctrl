package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bringer of the Black Dawn — Creature — Bringer {7}{B}{B}, 5/5:
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	 Trample
//	 At the beginning of your upkeep, you may pay 2 life. If you do,
//	 search your library for a card, then shuffle and put that card on
//	 top."
//
// The alternative cost is bringerAlternativeCost (CR 118.9). The upkeep
// ability is Erebos's "pay 2 life, then" asked as the trigger RESOLVES
// (MayChoice with LifeCost, CR 608.2d), so it carries no sandbox
// simplification: the payment is re-checked in the Yes branch (a
// player at 1 life cannot pay and searches for nothing), and only after
// it lands does the search start.
//
// The search is the one-mana tutors' own, tutorToTop: shuffle FIRST,
// then place the card on top, no reveal (the card prints none).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a9f98ad5-e8d4-4142-a19f-3368b85d1c95",
		Name:             "Bringer of the Black Dawn",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{bringerAlternativeCost()},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Bringer of the Black Dawn — you may pay 2 life to tutor", func(g *game.Game, item *game.StackItem) error {
				return MayChoice{
					Question: "Bringer of the Black Dawn — pay 2 life to search your library for a card?",
					LifeCost: 2,
					OnYes:    bringerOfTheBlackDawnPayAndSearch,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// bringerOfTheBlackDawnPayAndSearch is the Yes branch: pay 2 life, and
// only if that is possible, search for a card to put on top.
func bringerOfTheBlackDawnPayAndSearch(ctx *Context) error {
	controller := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Eliminated || p.Life < 2 {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), controller, 2); err != nil {
		return err
	}
	return tutorToTop(ctx, "Bringer of the Black Dawn — any card, to the top of your library", false, nil)
}
