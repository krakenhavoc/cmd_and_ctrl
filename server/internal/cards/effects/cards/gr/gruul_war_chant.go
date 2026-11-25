package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul War Chant — Enchantment {2}{R}{G} (Conspiracy: Take the Crown, S82
// older sets batch 273, #5540):
//
//	"Attacking creatures you control get +1/+0 and have menace."
//
// Goblin Oriflamme's layer 7c boost over the creatures you control while they
// attack, and Ohran Frostfang's layer 6 menace grant on the same set; both
// declare that they depend on attacking status, so a declaration, a creature
// put onto the battlefield attacking and the end of combat re-read them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9edb891-1a38-4b8a-8539-c459e8a52315",
		Name:         "Gruul War Chant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttackingCreaturesYouControlGetPower(1),
			effects.AttackingCreaturesHave("Attacking creatures you control have menace.", nil, "menace"),
		},
	})
}
