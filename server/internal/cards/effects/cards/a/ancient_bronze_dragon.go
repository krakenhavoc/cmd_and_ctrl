package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Bronze Dragon — {5}{G}{G} Creature — Elder Dragon 7/7
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, roll a d20. When
//	 you do, put X +1/+1 counters on each of up to two target creatures,
//	 where X is the result."
//
// "When you do" is a CR 603.12 reflexive trigger made as the roll resolves:
// its targets (up to two creatures, any controller) are chosen as it goes on
// the stack, and X is the die's result, carried as the trigger's amount.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ce74097-97b1-493d-9585-bd1336f95927",
		Name:            "Ancient Bronze Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Ancient Bronze Dragon — roll a d20; when you do, put X +1/+1 counters on each of up to two target creatures",
				effects.Do(effects.RollDice{Sides: 20, N: 1, Then: ancientBronzeDragonRolled})),
		},
	})
}

// ancientBronzeDragonRolled is "When you do": the reflexive trigger, told the
// result. An on-disk key: never renamed, never reused.
var ancientBronzeDragonRolled = effects.DiceTail("dice/ancient-bronze-dragon-roll", func(ctx *effects.Context, rolls []int) error {
	return effects.ReflexiveTrigger{
		Label:  "Ancient Bronze Dragon — put X +1/+1 counters on each of up to two target creatures",
		Body:   ancientBronzeDragonCounters,
		Params: game.EffectParams{Amount: rolls[0]},
	}.Apply(ctx)
})

var ancientBronzeDragonCounters = game.ReflexiveBody("gap-b215/ancient-bronze-dragon-counters",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: item.Params.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("up to two target creatures").WithCount(0, 2)
	}))
