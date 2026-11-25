package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain's Claws — Artifact — Equipment {2} (Oath of the Gatewatch, #5545):
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature attacks, create a 1/1 white Kor Ally creature
//	 token that's tapped and attacking.
//	 Equip {1}"
//
// The trigger belongs to the Claws (Moonsilver Spear's shape): its controller
// makes the token, which enters tapped and attacking the same player or
// planeswalker the equipped creature attacked (CR 508.4: it was never declared
// as an attacker, so no attack triggers fire).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cacd7de-83ac-4767-b6f2-f5a89f04a0b9",
		Name:         "Captain's Claws",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Captain's Claws — create a 1/1 white Kor Ally creature token that's tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					tmpl := effects.TokenCard("1/1 white Kor Ally")
					tmpl.Tapped = true
					tmpl.AttackingTarget = effects.TriggerAttackTarget(item)
					return effects.CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
