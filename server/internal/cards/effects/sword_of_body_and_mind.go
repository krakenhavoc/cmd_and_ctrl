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
// Declared simplification, weaker than printed (#259): PROTECTION is
// not modelled by the engine (#662), so the "protection from green and
// from blue" clause does nothing. The equipped creature can be blocked
// by green and blue creatures, targeted by green and blue spells, and
// takes damage from them normally. Everything else — the +2/+2, the
// Wolf, the ten-card mill and the equip cost — is exactly as printed.
func init() {
	Register(Spec{
		OracleID:     "fac42229-4f5f-4d04-85dd-5031d4e435aa",
		Name:         "Sword of Body and Mind",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The equipped creature does not get protection from green or from blue — green and blue removal, blockers and damage all affect it normally.",
		},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
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
