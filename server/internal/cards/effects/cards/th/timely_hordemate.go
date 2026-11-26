package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Timely Hordemate — Creature — Human Warrior {3}{W}, 3/2 (Khans of Tarkir):
//
//	"Raid — When this creature enters, if you attacked this turn, return
//	 target creature card with mana value 2 or less from your graveyard to
//	 the battlefield."
//
// An intervening "if" (CR 603.4) on Shepherd of the Cosmos's return: the
// target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b). It returns under your control, untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ace6510-7876-48e4-9cae-1c24f2229e03",
		Name:         "Timely Hordemate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB223RaidWhenThisEnters("Timely Hordemate — raid: return target creature card with mana value 2 or less from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFirstLegalGraveyardTargetToBattlefield(g, item)
				}),
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less in your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2))),
		},
	})
}
