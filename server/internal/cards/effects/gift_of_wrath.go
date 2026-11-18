package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9ba31dd0-d78e-4639-a550-bd72cac5898a",
		Name:         "Gift of Wrath",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, 2, 2),
			b43GrantToAttachedWhen(cmdr2B19AttachedIsCreature, "menace"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Gift of Wrath — create a 2/2 red Spirit creature token with menace",
				Do(CreateToken{Template: TokenCard("2/2 red Spirit with menace"), N: 1})),
		},
	})
}
