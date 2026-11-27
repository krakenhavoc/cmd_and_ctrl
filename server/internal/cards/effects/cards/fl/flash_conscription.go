package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flash Conscription — Instant {5}{R} (Ravnica: City of Guilds):
//
//	"Untap target creature and gain control of it until end of turn. That
//	 creature gains haste until end of turn. If {W} was spent to cast this spell,
//	 the creature gains "Whenever this creature deals combat damage, you gain that
//	 much life" until end of turn."
//
// Act of Treason's three clauses in printed order (the theft is a layer-2
// until-end-of-turn effect, so control reverts by itself at cleanup), then the
// white-mana rider: whether {W} was spent is read off the spell's own payment
// record as it resolves (CR 601.2h), and the granted trigger is Sunhome
// Enforcer's. "You" in the granted ability is whoever controls the creature when
// it triggers, which is the caster. A spell cast without paying (the engine
// knows nothing was spent) grants nothing, as printed.
//
// No simplification.
func init() {
	const grant = "flash-conscription/lifegain"
	effects.Register(effects.Spec{
		OracleID:     "a2019c23-09b8-4b0f-8b91-d8137804cb95",
		Name:         "Flash Conscription",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		// "If {W} was spent to cast this spell" reads its own payment, so a
		// waived one is known and not guessed (#3855).
		ReadsManaSpent: game.ReadsOwnPayment,
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Triggered: []game.TriggeredAbility{effects.ThisDealsCombatDamageYouGainThatMuchLife("Flash Conscription")},
			Text:      "Whenever this creature deals combat damage, you gain that much life.",
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Flash Conscription — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Flash Conscription — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if ctx.ManaSpentOfColor("W") == 0 {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target:   target,
				Keys:     []string{grant},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Flash Conscription — gains \"whenever this creature deals combat damage, you gain that much life\"",
			}.Apply(ctx)
		},
	})
}
