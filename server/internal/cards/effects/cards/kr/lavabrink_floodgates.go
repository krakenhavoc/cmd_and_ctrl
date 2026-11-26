package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavabrink Floodgates — {3}{R} Artifact (Commander 2020):
//
//	"{T}: Add {R}{R}.
//	 At the beginning of each player's upkeep, that player may put a doom
//	 counter on this artifact or remove a doom counter from it. Then if it
//	 has three or more doom counters on it, sacrifice this artifact. When you
//	 do, it deals 6 damage to each creature."
//
// Plague Boiler's put-or-remove choice, asked of the player whose upkeep it
// is (PickOption.Player), with "do nothing" first for the "may" and "remove"
// offered only while a counter is there. The count is checked after the
// choice whether or not they chose (the "Then if"). Sacrificing is a keyed
// tail; "when you do" is a reflexive trigger (CR 603.12), controlled by the
// Floodgates' controller, whose damage is dealt by the Floodgates as it last
// existed (CR 608.2h). Declared as a creature sweep: it is one.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd4df73c-bbd5-47fd-a961-19f3a3c59381",
		Name:         "Lavabrink Floodgates",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}{R}",
			Label:    "{T}: Add {R}{R}.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.AtEachUpkeep("Lavabrink Floodgates — that player may put a doom counter on it or remove one; then if it has three or more, sacrifice it",
				lavabrinkFloodgatesUpkeep),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}}),
		},
	})
}

func lavabrinkFloodgatesUpkeep(g *game.Game, item *game.StackItem) error {
	player := effects.TriggeringActor(item)
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !effects.OnBattlefield(g, item.SourceCardID) || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	options := []game.ChoiceOption{
		{Label: "Do nothing"},
		{Label: "Put a doom counter on Lavabrink Floodgates"},
	}
	if c.Counters["doom"] > 0 {
		options = append(options, game.ChoiceOption{Label: "Remove a doom counter from Lavabrink Floodgates"})
	}
	return effects.PickOption{
		Player:   player,
		Question: "Lavabrink Floodgates — put a doom counter on it, remove one, or neither?",
		Options:  options,
		ThenKey:  lavabrinkFloodgatesAnswer,
		WithItem: true,
	}.Apply(effects.NewContext(g, item))
}

var lavabrinkFloodgatesAnswer = effects.OptionPickThen("option-pick/lavabrink-floodgates-doom", func(ctx *effects.Context, r game.OptionPicked) error {
	switch r.Index {
	case 1:
		if err := effects.PutACounterOnThis("doom")(ctx.Game, ctx.Item); err != nil {
			return err
		}
	case 2:
		if err := effects.RemoveACounterFromThis("doom")(ctx.Game, ctx.Item); err != nil {
			return err
		}
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	if !ok || !effects.OnBattlefield(ctx.Game, ctx.Item.SourceCardID) || c.Counters["doom"] < 3 {
		return nil
	}
	return lavabrinkFloodgatesSacrifice(ctx.Game, ctx.Item)
})

var lavabrinkFloodgatesSacrifice = effects.SacrificeThisThen("then/lavabrink-floodgates-sacrificed", func(ctx *effects.Context) error {
	return effects.ReflexiveTrigger{
		Label: "Lavabrink Floodgates — it deals 6 damage to each creature",
		Body:  lavabrinkFloodgatesBlastBody,
	}.Apply(ctx)
})

var lavabrinkFloodgatesBlastBody = game.ReflexiveBody("lavabrink-floodgates/blast", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 6)
}), nil)
