package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Hospitality — Enchantment — Aura Curse {2}{R} (Innistrad:
// Crimson Vow):
//
//	"Enchant player
//	 Creatures attacking enchanted player have trample.
//	 Whenever a creature deals combat damage to enchanted player, that
//	 player exiles the top card of their library. Until end of turn, that
//	 creature's controller may play that card and they may spend mana as
//	 though it were mana of any color to cast that spell."
//
// The enchanted player is the Aura's AttachedTo (Curse of Opulence's
// shape). The static is a layer 6 grant to every creature declared as
// attacking that player, whoever controls it (a creature attacking a
// planeswalker they control is not attacking the player), rebuilt when an
// attacker is declared or leaves combat (DependsOnAttackingStatus).
//
// The trigger is one per creature that deals combat damage, controlled by
// the Curse's controller. The enchanted player exiles the top card of their
// own library, and the CREATURE's controller, not the Curse's, holds the
// this-turn permission to play it (a land counts against their land drop),
// with mana of any colour for a spell. The controller is read as the
// creature last existed (CR 608.2h), so a creature that died in combat
// still pays its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fb70b94-0847-4155-8953-cb53fef82ef1",
		Name:         "Curse of Hospitality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			Label:                    "Creatures attacking enchanted player have trample.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && source.AttachedTo.Kind == game.TargetPlayer &&
					target.AttackingTarget != uuid.Nil && target.AttackingTarget == source.AttachedTo.ID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"trample"})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !ev.Combat || ev.Amount <= 0 || source.AttachedTo.Kind != game.TargetPlayer || ev.Target != source.AttachedTo.ID {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.Source)
				return ok && src.IsCreature()
			}, "Curse of Hospitality — the damaged player exiles the top card of their library; the creature's controller may play it this turn",
				curseOfHospitalityExile),
		},
	})
}

// curseOfHospitalityExile is the trigger's body: the enchanted player
// exiles the top card of their library, and the controller of the creature
// that dealt the damage may play it this turn, spending mana as though it
// were of any colour.
func curseOfHospitalityExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ev := ctx.Trigger().Event
	// The damage event's actor is the creature's controller as it dealt the
	// damage (damage_tail.go), so a creature that has since died still pays
	// its controller (CR 608.2h).
	holder := ev.Actor
	if holder == uuid.Nil {
		creature, ok := g.LookupCardForEffect(ev.Source)
		if !ok {
			return nil
		}
		holder = creature.Controller
	}
	if holder == uuid.Nil || ev.Target == uuid.Nil {
		return nil
	}
	return effects.ExileTopWithPermission{From: ev.Target, GrantTo: holder, N: 1, AnyColor: true}.Apply(ctx)
}
