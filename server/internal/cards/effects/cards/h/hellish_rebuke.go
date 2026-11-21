package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellish Rebuke — Instant {2}{B} (Adventures in the Forgotten Realms
// Commander, Secret Lair):
//
//	"Until end of turn, permanents your opponents control gain 'When this
//	 permanent deals damage to the player who cast Hellish Rebuke,
//	 sacrifice this permanent. You lose 2 life.'"
//
// A duration grant of a catalog bundle (ADR 0093 PR 4): the permanents
// are the ones your opponents control as it resolves (CR 611.2c), so a
// permanent that arrives later has nothing. The granted trigger is the
// PERMANENT'S: its controller is whoever controls it when it triggers, so
// "you lose 2 life" is that player, and the sacrifice is theirs too. "The
// player who cast Hellish Rebuke" is the grant's "you" (the controller of
// the resolving spell), kept on the grant itself and judged per grant
// (TriggeredAbility.EventTargetIsGrantActivator), so a permanent given
// two Rebukes by two casters hurts only whoever it damages. Any damage
// counts, not just combat damage; damage to a planeswalker, a creature or
// to another player does not. The life is lost whether or not the
// permanent could be sacrificed.
//
// No simplification.
const hellishRebukeTrigger = "hellish-rebuke/damage-sacrifice"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d307515-6ea1-4b1f-8e4b-49bfebfbdfcc",
		Name:         "Hellish Rebuke",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: hellishRebukeTrigger,
			Triggered: []game.TriggeredAbility{
				hellishRebukeDamageTrigger(),
			},
			Text: "When this permanent deals damage to the player who cast Hellish Rebuke, sacrifice this permanent. You lose 2 life.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.OpponentControls(),
				Keys:  []string{hellishRebukeTrigger},
				Label: "Hellish Rebuke — permanents your opponents control gain a damage trigger until end of turn",
			}.Apply(ctx)
		},
	})
}

func hellishRebukeDamageTrigger() game.TriggeredAbility {
	t := effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return source != nil && ev.Kind == game.EventDealDamage && ev.Amount > 0 &&
			ev.Source == source.InstanceID && g.PlayerByIDForEffect(ev.Target) != nil
	}, "Hellish Rebuke — sacrifice this permanent and lose 2 life",
		effects.Do(effects.Instruction(func(ctx *effects.Context) error {
			return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
		}), effects.YouLoseLife(2)))
	t.EventTargetIsGrantActivator = true
	return t
}
