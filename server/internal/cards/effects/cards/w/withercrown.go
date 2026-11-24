package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withercrown — Enchantment — Aura {1}{B} (Kaldheim, #5505):
//
//	"Enchant creature
//	 Enchanted creature has base power 0 and has "At the beginning of
//	 your upkeep, you lose 1 life unless you sacrifice this creature.""
//
// Base power is layer 7b (CR 613.4b); the upkeep trigger is a granted
// ability, so "your" and "this creature" are the enchanted creature's
// controller and the creature (Morgul-Knife Wound's shape).
const withercrownGrant = "withercrown/lose-life-unless-sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc700a38-7467-41fd-861f-940e0e5982bd",
		Name:         "Withercrown",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: withercrownGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Withercrown — you lose 1 life unless you sacrifice this creature", withercrownUpkeep),
			},
			Text: "At the beginning of your upkeep, you lose 1 life unless you sacrifice this creature.",
		}},
		Static: []game.StaticAbility{
			basePowerZeroAttached(),
			effects.GrantAbilitiesToAttached(withercrownGrant),
		},
	})
}

func basePowerZeroAttached() game.StaticAbility {
	s := effects.SetAttachedBasePT(0, 0)
	s.Label = "Enchanted creature has base power 0."
	s.Apply = func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power = 0 }
	return s
}

func withercrownUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	return effects.PickOption{
		Question: "Withercrown — sacrifice this creature, or lose 1 life",
		Options:  []game.ChoiceOption{{Label: "Sacrifice this creature"}, {Label: "Lose 1 life"}},
		ThenKey:  withercrownChose,
		WithItem: true,
	}.Apply(ctx)
}

var withercrownChose = effects.OptionPickThen("option-pick/withercrown-sacrifice-or-lose-life", func(ctx *effects.Context, r game.OptionPicked) error {
	item := ctx.Item
	if r.Option != nil && r.Option.Label == "Sacrifice this creature" {
		if !effects.OnBattlefield(ctx.Game, item.SourceCardID) || effects.SourceIsNewObject(ctx.Game, item) {
			return nil
		}
		return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
})
