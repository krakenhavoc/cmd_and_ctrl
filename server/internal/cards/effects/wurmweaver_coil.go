package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wurmweaver Coil — Enchantment — Aura {4}{G}{G} (Ravnica Remastered):
//
//	"Enchant green creature
//	 Enchanted creature gets +6/+6.
//	 {G}{G}{G}, Sacrifice this Aura: Create a 6/6 green Wurm creature token."
//
// The Aura is the source of the activated ability, so the token is made
// even if the host has since left (the Aura would already be gone then).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83419229-c266-4022-8297-d1b05ec5ee20",
		Name:         "Wurmweaver Coil",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(OfColor("green")),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(6, 6)},
		Activated: []ActivatedAbility{{
			Label:   "{G}{G}{G}, Sacrifice this Aura: Create a 6/6 green Wurm creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeSelf: true, Mana: "{G}{G}{G}"},
			Effect:  Do(CreateToken{Template: TokenCard("6/6 green Wurm"), N: 1}),
		}},
	})
}
