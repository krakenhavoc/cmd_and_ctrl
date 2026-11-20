package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitesail Apprentice — Creature — Kor Soldier {W}, 1/1 (Worldwake, #5582):
//
//	"As long as this creature is equipped, it gets +1/+1 and has flying."
//
// Armed Assailant's shape: a layer-7c pump and a layer-6 grant, both
// while an Equipment is attached to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2157f2c-a3a4-47aa-b33e-b2c65e11c2e5",
		Name:         "Kitesail Apprentice",
		Completeness: CompletenessFull,
		Static:       kitesailApprenticeStatics(),
	})
}

func kitesailApprenticeStatics() []game.StaticAbility {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}
	pump := b16Anthem(whileEquipped, 1, 1)
	pump.Label = "As long as this creature is equipped, it gets +1/+1."
	flying := KeywordGrant(whileEquipped, "flying")
	flying.Label = "As long as this creature is equipped, it has flying."
	return []game.StaticAbility{pump, flying}
}
