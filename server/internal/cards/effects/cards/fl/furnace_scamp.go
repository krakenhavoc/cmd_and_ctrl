package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furnace Scamp — Creature — Phyrexian Beast {R}, 1/1 (New Phyrexia, #5575):
//
//	"Whenever this creature deals combat damage to a player, you may sacrifice
//	 it. If you do, this creature deals 3 damage to that player."
//
// The "you may" is the controller's choice as the trigger resolves, and "if you
// do" hangs off the Scamp really leaving the battlefield; the 3 damage is dealt
// by the Scamp as it last existed (CR 608.2h) to the player it just damaged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0868a52-5b19-4429-acf1-399f3ae8308c",
		Name:         "Furnace Scamp",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Furnace Scamp — you may sacrifice it; if you do, it deals 3 damage to that player",
				effects.Do(effects.MayChoice{
					Question: "Furnace Scamp — sacrifice it to deal 3 damage to that player?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     furnaceScampAnswer,
				})),
		},
	})
}

// furnaceScampAnswer is the "yes": sacrifice the Scamp, then the damage once
// it has gone. An on-disk key.
var furnaceScampAnswer = effects.MayYes("confirm/furnace-scamp-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ItemTail(furnaceScampDamage)}.Apply(ctx)
})

// furnaceScampDamage is "this creature deals 3 damage to that player", run once
// the Scamp is gone. An on-disk key.
var furnaceScampDamage = effects.CardTail("then/furnace-scamp-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: ctx.Item.Trigger.Event.Target, Amount: 3}.Apply(ctx)
})
