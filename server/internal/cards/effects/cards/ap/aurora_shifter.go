package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aurora Shifter — Creature — Shapeshifter {1}{U}, 1/3 (Modern Horizons 3
// Commander, #5948):
//
//	"Whenever this creature deals combat damage to a player, you get that
//	 many {E}.
//	 At the beginning of combat on your turn, you may pay {E}{E}. When you
//	 do, this creature becomes a copy of another target creature you
//	 control, except it has this ability and "Whenever this creature deals
//	 combat damage to a player, you get that many {E}.""
//
// The second line is Saheeli, Radiant Creator's shape: the payment is made
// as the combat trigger resolves and the copy's target is chosen as the
// reflexive trigger (CR 603.12) goes on the stack. The copy has no stated
// duration (CR 611.2a), as The Ever-Changing 'Dane's. "Except it has this
// ability and ..." (CR 707.9a) is one grant of both triggers as a named
// bundle, so a Shifter that has become a Bear still gets energy and can
// copy again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d499e473-aed8-4c8b-8cf0-a43bcc5f40cf",
		Name:         "Aurora Shifter",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{auroraShifterDamageTrigger(), auroraShifterCombatTrigger()},
		Grants: []effects.AbilityGrant{{
			Key:       auroraShifterGrant,
			Triggered: []game.TriggeredAbility{auroraShifterDamageTrigger(), auroraShifterCombatTrigger()},
			Text:      "Whenever this creature deals combat damage to a player, you get that many {E}. At the beginning of combat on your turn, you may pay {E}{E}. When you do, this creature becomes a copy of another target creature you control, except it has this ability and \"Whenever this creature deals combat damage to a player, you get that many {E}.\"",
		}},
	})
}

const auroraShifterGrant = "aurora-shifter/this-ability"

func auroraShifterDamageTrigger() game.TriggeredAbility {
	return effects.WheneverThisDealsCombatDamageToAPlayer("Aurora Shifter — you get that many {E}",
		func(g *game.Game, item *game.StackItem) error {
			return effects.GetEnergy{N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
		})
}

func auroraShifterCombatTrigger() game.TriggeredAbility {
	return effects.AtBeginningOfYourCombat("Aurora Shifter — you may pay {E}{E}",
		effects.MayPayEnergyThen("Aurora Shifter", 2, "copy another creature you control", auroraShifterPaid))
}

// auroraShifterPaid is "you may pay {E}{E}. When you do": the energy was
// paid, so the reflexive trigger is created. An on-disk key.
var auroraShifterPaid = effects.EnergyPaid("pay-unless/aurora-shifter-copy",
	func(g *game.Game, item *game.StackItem) error {
		return effects.WhenYouDo("Aurora Shifter — this creature becomes a copy of another target creature you control",
			auroraShifterCopyBody).Apply(effects.NewContext(g, item))
	})

// auroraShifterCopyBody is the reflexive trigger's body. An on-disk key.
var auroraShifterCopyBody = game.ReflexiveBody("aurora-shifter/copy",
	effects.SimpleBody(auroraShifterBecomeCopy),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	}))

func auroraShifterBecomeCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var of uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			of = t.ID
			break
		}
	}
	if of == uuid.Nil || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.BecomeCopy{
		Targets:  []uuid.UUID{item.SourceCardID},
		Of:       of,
		Duration: effects.CopyIndefinite,
		Except:   func(v *game.PrintedValues) { v.GrantAbility(auroraShifterGrant) },
		Label:    "Aurora Shifter — becomes a copy of another creature you control",
	}.Apply(ctx)
}
