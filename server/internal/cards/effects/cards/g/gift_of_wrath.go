package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Wrath — Enchantment — Aura {3}{R} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets +2/+2 and has
//	 menace.
//	 When this Aura leaves the battlefield, create a 2/2 red Spirit
//	 creature token with menace."
//
// The pump and menace share Careful Cultivation's condition. The token is
// made on any exit (the host dying, a Disenchant, the Aura fizzling off a
// host that stopped being legal), by the Aura's controller as it left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ba31dd0-d78e-4639-a550-bd72cac5898a",
		Name:         "Gift of Wrath",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, 2, 2),
			effects.B43GrantToAttachedWhen(effects.Cmdr2B19AttachedIsCreature, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Gift of Wrath — create a 2/2 red Spirit creature token with menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Spirit with menace"), N: 1})),
		},
	})
}
