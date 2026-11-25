package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of Keys — Creature — Human Rogue Mutant {3}{U}{U}, 4/4 (Conspiracy:
// Take the Crown, S82 older sets batch 272, #5540):
//
//	"When this creature enters, you become the monarch.
//	 At the beginning of your upkeep, if you're the monarch, creatures you
//	 control can't be blocked this turn."
//
// The monarch ETB is the shared one (WhenThisEntersYouBecomeTheMonarch).
// The upkeep trigger has an intervening "if" (CR 603.4), checked as the
// upkeep begins and again as the ability resolves; its effect is Tezzeret's
// Gatebreaker's scoped "creatures you control can't be blocked" until end of
// turn, which covers a creature that enters later this turn too (CR 611.2c
// does not apply: it changes no characteristic).
//
// No simplification.
func init() {
	const label = "Keeper of Keys — creatures you control can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "1fa0f877-e26c-4a26-8c69-4bf8fcdd30b5",
		Name:         "Keeper of Keys",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Keeper of Keys"),
			effects.InterveningIf(effects.CondMonarch{},
				effects.AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
					return effects.RestrictUntilEOT{Scope: game.ScopeYourCreatures, Restrictions: game.CantBeBlocked, Label: label}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
