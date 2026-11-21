package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paradox Haze — Enchantment — Aura {2}{U} (Special Guests):
//
//	"Enchant player
//	 At the beginning of enchanted player's first upkeep each turn, that
//	 player gets an additional upkeep step after this step."
//
// "First upkeep each turn" is the step ordinal (the added upkeeps are
// later ones, so the Haze does not feed on itself). The extra upkeep is
// CR 500.10: a beginning phase of its own holding only the upkeep step,
// and it is added only on that player's own turn (CR 500.10a, "that
// player gets").
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daff0e7b-08f2-435f-97b9-9b000e94ffa8",
		Name:         "Paradox Haze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && ev.Actor == source.AttachedTo.ID &&
					g.IsFirstStepOfItsKindForEffect()
			}, "Paradox Haze — that player gets an additional upkeep step after this step",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					player := item.Trigger.Event.Actor
					g.AddUpkeepStepsAfterThisPhaseForEffect(item.SourceCardID, player, 1)
					return nil
				}),
		},
	})
}
