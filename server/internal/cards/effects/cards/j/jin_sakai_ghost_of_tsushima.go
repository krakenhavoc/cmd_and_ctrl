package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jin Sakai, Ghost of Tsushima — Legendary Creature — Human Samurai
// {1}{W}{U}{B}, 2/4 (Secret Lair Drop):
//
//	"Whenever Jin Sakai deals combat damage to a player, draw a card.
//	 Whenever a creature you control attacks a player, if no other creatures
//	 are attacking that player, choose one —
//	 • Standoff — It gains double strike until end of turn.
//	 • Ghost — It can't be blocked this turn."
//
// The second ability triggers once per attacking creature you control that was
// declared against a player; its intervening "if" (CR 603.4) counts the other
// creatures attacking that same player, whoever controls them, as the trigger
// goes on the stack and again as it resolves. "It" is the attacker the event
// names; if it has left the battlefield by then, nothing happens.
//
// No simplification.
func init() {
	attacks := effects.On(game.EventAttack, jinSakaiLoneAttacker, "Jin Sakai — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	attacks.Modes = effects.ChooseOne(
		effects.ModeDoing("Standoff — It gains double strike until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				id, ok := jinSakaiAttacker(ctx, item)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"},
					Label: "Jin Sakai — Standoff: double strike"}.Apply(ctx)
			}),
		effects.ModeDoing("Ghost — It can't be blocked this turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				id, ok := jinSakaiAttacker(ctx, item)
				if !ok {
					return nil
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
					Label: "Jin Sakai — Ghost: can't be blocked"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "76c43ac8-a878-4239-b0e3-f0254ef63e50",
		Name:         "Jin Sakai, Ghost of Tsushima",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Jin Sakai — draw a card", effects.Do(effects.DrawCards{N: 1})),
			attacks,
		},
	})
}

// jinSakaiLoneAttacker is the trigger condition: a creature its controller
// controls was declared as an attacker against a player, and no other creature
// is attacking that player.
func jinSakaiLoneAttacker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor != source.Controller {
		return false
	}
	return jinSakaiAloneAgainst(g, ev.CardID, ev.Target)
}

// jinSakaiAloneAgainst is true when `attacker` is attacking the player
// `target` and no other creature is.
func jinSakaiAloneAgainst(g *game.Game, attacker, target uuid.UUID) bool {
	if g.ClassifyAttackTargetForEffect(target) != game.AttackTargetPlayer {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != attacker && c.IsCreature() && c.AttackingTarget == target {
			return false
		}
	}
	return true
}

// jinSakaiAttacker is the creature the trigger is about, if it is still the
// lone attacker of that player as the trigger resolves.
func jinSakaiAttacker(ctx *effects.Context, item *game.StackItem) (uuid.UUID, bool) {
	if item.Trigger == nil {
		return uuid.Nil, false
	}
	ev := item.Trigger.Event
	if !effects.OnBattlefield(ctx.Game, ev.CardID) || !jinSakaiAloneAgainst(ctx.Game, ev.CardID, ev.Target) {
		return uuid.Nil, false
	}
	return ev.CardID, true
}
