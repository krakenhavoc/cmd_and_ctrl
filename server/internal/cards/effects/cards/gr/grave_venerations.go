package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grave Venerations — Enchantment {3}{B} (Lorwyn Eclipsed Commander,
// #5368):
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your end step, if you're the monarch, return up to
//	 one target creature card from your graveyard to your hand.
//	 Whenever a creature you control dies, each opponent loses 1 life and
//	 you gain 1 life."
//
// The end-step trigger has an intervening "if" (CR 603.4): it must be the
// monarch as the end step begins and again as the trigger resolves (the
// condition is declared with the trigger, so a crown lost in between does
// nothing). The target is "up to one", so no target is a legal choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cb71772-592d-4dd1-91af-816933af760c",
		Name:         "Grave Venerations",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Grave Venerations"),
			effects.Targeting(effects.InterveningIf(effects.CondMonarch{},
				effects.On(game.EventBeginEndStep,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.ByYou(ev, source, game.Characteristic{}, g) && effects.YoureTheMonarch(g, source.Controller)
					},
					"Grave Venerations — return up to one target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						if !effects.YoureTheMonarch(g, item.Controller) {
							return nil
						}
						return effects.ReturnTargetCardToHand(g, item)
					})),
				effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1)),
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Grave Venerations — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
