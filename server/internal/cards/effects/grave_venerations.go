package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1cb71772-592d-4dd1-91af-816933af760c",
		Name:         "Grave Venerations",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Grave Venerations"),
			Targeting(InterveningIf(CondMonarch{},
				On(game.EventBeginEndStep,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return ByYou(ev, source, game.Characteristic{}, g) && YoureTheMonarch(g, source.Controller)
					},
					"Grave Venerations — return up to one target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						if !YoureTheMonarch(g, item.Controller) {
							return nil
						}
						return returnTargetCardToHand(g, item)
					})),
				TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1)),
			TriggerWithPurpose(WheneverACreatureYouControlDies("Grave Venerations — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
