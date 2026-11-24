package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fires of Mount Doom — Legendary Enchantment {2}{R} (The Lord of the
// Rings: Tales of Middle-earth, #5948 GAP2 b07):
//
//	"When Fires of Mount Doom enters, it deals 2 damage to target
//	 creature an opponent controls. Destroy all Equipment attached to
//	 that creature.
//	 {2}{R}: Exile the top card of your library. You may play that card
//	 this turn. When you play a card this way, Fires of Mount Doom deals
//	 2 damage to each player."
//
// The enters trigger deals the damage, then destroys every Equipment
// still attached to the target (Awaken the Sleeper's sweep), whether or
// not the damage killed it: an Equipment that fell off when its creature
// died is no longer attached, as printed.
//
// The activation is an impulse exile (ExileTopWithPermissionForEffect):
// "play", so a land can be played as the turn's land drop. Its
// permission's follow-up (FollowUp, with FollowUpOnLandPlay for a land)
// runs exactly when the card is played BECAUSE of this permission, and
// creates the delayed trigger "when you play a card this way" (CR
// 603.7a, the #6091 seam QueueDelayedTriggerForPermissionPlayForEffect):
// its source is Fires of Mount Doom, read by last-known information if
// it has left (CR 603.7e, 608.2h), and its controller is the player who
// activated. It deals 2 damage to each player.
//
// No simplification.
func init() {
	const name = "Fires of Mount Doom"
	effects.Register(effects.Spec{
		OracleID:     "f4944fde-2066-4aed-85c0-6e5e00a4bfe6",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(name+" — 2 damage to target creature an opponent controls; destroy all Equipment attached to it",
				firesOfMountDoomEnters),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Exile the top card of your library. You may play that card this turn. When you play a card this way, Fires of Mount Doom deals 2 damage to each player.",
			Purpose: game.Purpose{Impulse: 1, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref, _ := g.SourceObjectForEffect(item)
				_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{
					Source:             ref.ID,
					SourceName:         name,
					Label:              name + " — you may play that card this turn",
					FollowUp:           firesOfMountDoomPlayed,
					FollowUpOnLandPlay: true,
					FollowUpObject:     &ref,
				})
				return err
			},
		}},
	})
}

// firesOfMountDoomEnters deals the 2 damage, then destroys every
// Equipment attached to the target.
func firesOfMountDoomEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil // the target is gone: the ability does nothing (CR 608.2b)
	}
	creature := legal[0].ID
	if err := effects.DealDamageToTheTarget(2)(g, item); err != nil {
		return err
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasSubtype("Equipment") && c.IsAttachedTo(creature) {
			if err := (effects.DestroyTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
	}
	return nil
}

// firesOfMountDoomPlayed creates the delayed trigger when the card is
// played through the permission. An on-disk key.
var firesOfMountDoomPlayed = game.RegisterCastFollowUp("fires-of-mount-doom-played",
	func(g *game.Game, f game.CastFollowUp) error {
		g.QueueDelayedTriggerForPermissionPlayForEffect(f, f.Player, game.ReflexiveTrigger{
			Label: "Fires of Mount Doom — 2 damage to each player",
			Body:  firesOfMountDoomDamageBody,
		})
		return nil
	})

// firesOfMountDoomDamageBody deals 2 damage to each player, from Fires of
// Mount Doom. An on-disk key.
var firesOfMountDoomDamageBody = game.ReflexiveBody("fires-of-mount-doom/damage-each-player",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		return effects.B38DamageEachPlayer(effects.NewContext(g, item), 2)
	}, nil)
