package cardsso

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Tithe — Enchantment — Aura {1}{W} (Return to Ravnica, S82 OLD):
//
//	"Enchant nonland permanent
//	 At the beginning of the upkeep of enchanted permanent's controller, that
//	 player sacrifices it unless they pay {X}, where X is its mana value."
//
// Dance of the Dead's upkeep trigger (UpkeepOfEnchantedCreaturesController,
// which reads any host) asks the upkeep's player, as it resolves (CR 118.12).
// X is the permanent's mana value as the trigger resolves; a permanent
// worth {0} costs nothing and stays. Declining, or being unable to pay,
// sacrifices the permanent, but only while it is still on the battlefield
// under that player's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60ff5af4-eb95-4544-9f63-d42c22112486",
		Name:         "Soul Tithe",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Nonland()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Soul Tithe — that player sacrifices it unless they pay {X}, where X is its mana value",
				soulTitheUpkeep),
		},
	})
}

func soulTitheUpkeep(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	host := effects.AttachedHostFor(g, item.SourceCardID)
	if host == nil || !effects.OnBattlefield(g, host.InstanceID) {
		return nil
	}
	mv := host.ManaValue()
	if mv <= 0 {
		return nil
	}
	return effects.PayUnless{
		Chooser:  item.Trigger.Event.Actor,
		Cost:     "{" + strconv.Itoa(mv) + "}",
		Question: "Soul Tithe — pay {" + strconv.Itoa(mv) + "} or sacrifice the enchanted permanent?",
		Else:     soulTitheDeclined,
		Carry:    []uuid.UUID{host.InstanceID},
	}.Apply(ctx)
}

// soulTitheDeclined is "that player sacrifices it" once the tax is not paid,
// keyed (#4256): the permanent rides the prompt as its Carry. An on-disk key:
// never renamed, never reused.
var soulTitheDeclined = effects.PayUnlessElse("pay-unless/soul-tithe-sacrifice", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return effects.SacrificePermanent{Target: r.Carry[0]}.Apply(ctx)
})
