package cardsh

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heroic Feast — {2}{G} Enchantment (Marvel Super Heroes, #4455):
//
//	"When this enchantment enters, create a Food token.
//	 Whenever you gain life, choose up to that many target creatures you
//	 control. Put a +1/+1 counter on each of them."
//
// The life-gain trigger reads the amount gained from its event, so the
// target clause is built per trigger (TargetsFrom): up to that many
// creatures you control, chosen as the trigger goes on the stack
// (CR 603.3d), and each still-legal one gets a counter as it resolves.
// One trigger per life-gain event, so a lifelink hit for 3 is one trigger
// with up to three targets.
//
// No simplification.
func init() {
	feast := effects.WheneverYouGainLife("Heroic Feast — put a +1/+1 counter on each of up to that many target creatures you control",
		heroicFeastCounters)
	feast.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		n := tc.Event.Amount
		return effects.TargetCreature("up to "+strconv.Itoa(n)+" target creatures you control", effects.YouControl()).WithCount(0, n)
	}
	effects.Register(effects.Spec{
		OracleID:     "31cae00a-3a74-4a59-80b0-a656f524e878",
		Name:         "Heroic Feast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heroic Feast — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			feast,
		},
	})
}

// heroicFeastCounters puts a +1/+1 counter on each target still legal.
func heroicFeastCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
