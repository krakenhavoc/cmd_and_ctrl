package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardians of Oboro — Creature — Moonfolk Samurai {2}{U}, 3/4 (Kamigawa:
// Neon Dynasty):
//
//	"Defender
//	 Modified creatures you control can attack as though they didn't have
//	 defender. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// A layer 6 static writing an unrestricted DefenderWaiver on each modified
// creature you control (Bristlepack Sentry's seam), read live at declaration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42b28dfb-de4d-4005-b0d7-1db3917cf159",
		Name:            "Guardians of Oboro",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static:          []game.StaticAbility{guardiansOfOboroWaiver()},
	})
}

// guardiansOfOboroWaiver is Bristlepack Sentry's waiver static, aimed at each
// modified creature its controller controls instead of at itself.
func guardiansOfOboroWaiver() game.StaticAbility {
	ab := effects.AvB18CanAttackAsThoughNoDefenderWhile(
		"Modified creatures you control can attack as though they didn't have defender.",
		func(*game.Game, *game.Card) bool { return true })
	ab.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && effects.B07IsModified(g, *target)
	}
	return ab
}
