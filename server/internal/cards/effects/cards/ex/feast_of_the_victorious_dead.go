package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast of the Victorious Dead — Enchantment {W}{B} (March of the Machine:
// The Aftermath, #5476):
//
//	"At the beginning of your end step, if one or more creatures died this
//	 turn, you gain that much life and distribute that many +1/+1 counters
//	 among any number of creatures you control."
//
// The intervening "if" is checked as the trigger would fire and again as it
// resolves (CR 603.4): the count of creatures that died this turn, any
// controller's, tokens included (b11CreaturesDiedThisTurn). "That much" is
// that count read as the trigger resolves. Nothing is targeted: the
// controller chooses the creatures and the split as it resolves
// (matDistributePlusOneCounters).
//
// No simplification.
func init() {
	trigger := effects.AtYourEndStep("Feast of the Victorious Dead — gain that much life and distribute that many +1/+1 counters among any number of creatures you control",
		func(g *game.Game, item *game.StackItem) error {
			n := effects.B11CreaturesDiedThisTurn(g)
			if n <= 0 {
				return nil
			}
			ctx := effects.NewContext(g, item)
			if err := (effects.GainLife{Player: item.Controller, Amount: n}).Apply(ctx); err != nil {
				return err
			}
			return effects.MatDistributePlusOneCounters(ctx, n)
		})
	trigger.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return effects.ByYou(ev, source, ch, g) && effects.B11CreaturesDiedThisTurn(g) > 0
	}
	effects.Register(effects.Spec{
		OracleID:     "7e4e1d13-db1a-48e5-9b5f-f3e71d10c23e",
		Name:         "Feast of the Victorious Dead",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
