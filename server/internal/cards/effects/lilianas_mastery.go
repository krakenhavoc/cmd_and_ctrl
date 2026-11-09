package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Mastery — Enchantment {3}{B}{B} (Foundations Commander):
//
//	"Zombies you control get +1/+1.
//	 When this enchantment enters, create two 2/2 black Zombie creature
//	 tokens."
//
// A layer 7c anthem over the controller's Zombies (effective subtypes, so a
// changeling counts; TribalAnthem is the shared lord body) and an enters
// trigger for two 2/2 black Zombie tokens, which the anthem pumps to 3/3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd104c7e-311e-4b03-98d3-5f20f3a99d26",
		Name:         "Liliana's Mastery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Liliana's Mastery — create two 2/2 black Zombie tokens",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 2})),
		},
	})
}
