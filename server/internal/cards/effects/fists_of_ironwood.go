package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fists of Ironwood — Enchantment — Aura {1}{G} (Ravnica Remastered):
//
//	"Enchant creature
//	 When this Aura enters, create two 1/1 green Saproling creature tokens.
//	 Enchanted creature has trample."
//
// The trample is a layer 6 grant on the host; the enters trigger makes the
// tokens whether or not the host is still there (they are not tied to it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f6a5278-23bf-47b2-a2b9-dfff7d96f3dd",
		Name:         "Fists of Ironwood",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fists of Ironwood — create two 1/1 green Saproling creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 2})),
		},
	})
}
