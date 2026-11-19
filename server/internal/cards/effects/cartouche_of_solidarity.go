package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cartouche of Solidarity — {W} Enchantment — Aura Cartouche: "Enchant
// creature you control. When this Aura enters, create a 1/1 white
// Warrior creature token with vigilance. Enchanted creature gets +1/+1
// and has first strike."
func init() {
	Register(Spec{
		OracleID:     "b9176173-c4a1-45ae-8f34-bd6b31c141fe",
		Name:         "Cartouche of Solidarity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cartouche of Solidarity — create a 1/1 white Warrior token with vigilance",
				Do(CreateToken{Template: TokenCard("1/1 white Warrior with vigilance"), N: 1})),
		},
	})
}
