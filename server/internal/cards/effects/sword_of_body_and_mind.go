package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of Body and Mind — Artifact — Equipment {3} (EDHREC rank
// 4320):
//
//	"Equipped creature gets +2/+2 and has protection from green and
//	 from blue.
//	 Whenever equipped creature deals combat damage to a player, you
//	 create a 2/2 green Wolf creature token and that player mills ten
//	 cards.
//	 Equip {2}"
//
// The first and weakest of the Mirrodin sword cycle, and the one a
// Commander deck plays for the mill rather than the protection: ten
// cards a hit is a real clock in a format where a self-mill deck is
// across the table and a real threat in one where nobody is.
//
// The trigger is the Swords' shared shape — EventDealDamage with
// ev.Combat, matched through the attachment rather than through a
// controller — and its two halves belong to different players: YOU get
// the Wolf, THAT PLAYER mills. The mill goes to the damaged player, who
// is read off the event rather than assumed to be the defending player
// of the combat, because a redirect could make them differ.
//
// PROTECTION FROM GREEN AND FROM BLUE is a layer-6 grant to the
// equipped creature (GrantToAttached), enforced since #662 /
// ADR 0072: a green or blue spell cannot target the wearer
// (CR 702.16b), a green or blue Aura falls off it and a green or blue
// Equipment unattaches (CR 702.16c-d), green or blue damage to it is
// prevented (CR 702.16e), and a green or blue creature cannot block
// it (CR 702.16f). Sword of Feast and Famine carries the identical
// clause, just against different colours.
func init() {
	Register(Spec{
		OracleID:     "fac42229-4f5f-4d04-85dd-5031d4e435aa",
		Name:         "Sword of Body and Mind",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("protection from green", "protection from blue"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Only the damaged player's ID is captured, per the
				// closure contract on every trigger Build.
				victim := ev.Target
				return game.NewTriggeredItem(source, "Sword of Body and Mind — a 2/2 Wolf, and that player mills ten",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (CreateToken{
							Controller: item.Controller,
							Template:   TokenCard("2/2 green Wolf"),
							N:          1,
						}).Apply(ctx); err != nil {
							return err
						}
						return MillCards{Player: victim, N: 10}.Apply(ctx)
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
