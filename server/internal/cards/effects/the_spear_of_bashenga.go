package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Spear of Bashenga — Legendary Artifact — Equipment {4}{W} (Marvel
// Super Heroes Commander, #4455):
//
//	"When The Spear of Bashenga enters, if there is no monarch, you become
//	 the monarch.
//	 Equipped creature gets +2/+2 and has vigilance.
//	 Whenever equipped creature attacks the monarch, destroy target tapped
//	 nonland permanent that player controls.
//	 Equip {2}"
//
// Three printed abilities beside the static and Equip, two triggers
// registered.
//
//   - The enters trigger carries its intervening "if" (CR 603.4): there is
//     no monarch when the Spear enters, and again as the trigger resolves
//     (CondBoard), so a crown handed out in response is not taken.
//   - "Attacks the monarch" is the attack event of the attached creature
//     (Captain America's Shield's condition) aimed at the player who holds
//     the crown right now: a planeswalker or battle that the monarch
//     controls is not "the monarch", and an attack on anyone else does
//     nothing. The target clause is built as the trigger goes on the stack
//     and names the player attacked, so "that player" is the monarch the
//     attacker was declared against; a permanent that is untapped, a land,
//     or changed hands in response is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7a0b370-913d-4e96-8fb4-5df168405073",
		Name:         "The Spear of Bashenga",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "there is no monarch",
				Test: func(ctx *Context) bool { return ctx.Game.Monarch == uuid.Nil },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && g.Monarch == uuid.Nil
			}, "The Spear of Bashenga — you become the monarch", Do(BecomeTheMonarch{}))),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureAttacked(ev, source) && g.Monarch != uuid.Nil && ev.Target == g.Monarch
				},
				TargetsFrom: spearOfBashengaTarget,
				Key:         "The Spear of Bashenga — destroy target tapped nonland permanent the monarch controls",
				Effect:      destroyTheFirstLegalTarget,
			},
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// spearOfBashengaTarget is "target tapped nonland permanent that player
// controls": the player the equipped creature attacked, off the event.
func spearOfBashengaTarget(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return TargetPermanent("target tapped nonland permanent that player controls",
		ControlledBy(tc.Event.Target), Nonland(), b751Tapped())
}
