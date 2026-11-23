package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eastfarthing Farmer — Creature — Halfling Peasant {2}{W}, 2/3 (The Lord of
// the Rings: Tales of Middle-earth Commander):
//
//	"When this creature enters, create a Food token. When you do, target
//	 creature you control gets +1/+1 until end of turn for each Food you
//	 control."
//
// Bag End Banquet's Food, then a CR 603.12 reflexive trigger (Gimli's shape)
// whose target ("target creature you control") is chosen as it goes on the
// stack. The number of Foods is counted as it resolves, so the new Food and
// any Food sacrificed in response are accounted for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b1603f3-a246-449b-a109-99adb49e5ea4",
		Name:         "Eastfarthing Farmer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eastfarthing Farmer — create a Food token; when you do, target creature you control gets +1/+1 until end of turn for each Food you control",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}, effects.Instruction(func(ctx *effects.Context) error {
					return effects.ReflexiveTrigger{
						Label: "Eastfarthing Farmer — target creature you control gets +1/+1 until end of turn for each Food you control",
						Cards: []uuid.UUID{ctx.Source()},
						Body:  eastfarthingFarmerPumpBody,
					}.Apply(ctx)
				}))),
		},
	})
}

var eastfarthingFarmerPumpBody = game.ReflexiveBody("eastfarthing-farmer/pump", effects.SimpleBody(eastfarthingFarmerPump),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))

func eastfarthingFarmerPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Food") })
	if n <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Eastfarthing Farmer — +1/+1 for each Food"}.Apply(ctx)
}
