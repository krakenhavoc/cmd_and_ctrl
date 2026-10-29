package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radagast of Rhosgobel — Legendary Creature — Avatar Wizard {2}{G}{G}, 2/5:
//
//	"The first creature spell you cast each turn costs {2} less to cast
//	 and can be cast as though it had flash."
//
// Shadow in the Warp's discount (generic mana only, priced before the
// spell is counted, so "no creature spell cast yet" means "this one is
// the first") and Yeva's flash grant, both gated on the same per-turn
// tally. The flash half is a derived cast-timing declaration
// (CastTimingRule.FirstCreatureSpellThisTurn), so it stops when Radagast
// leaves and the second creature spell of the turn finds the window shut.
//
// No simplification.
func init() {
	const label = "The first creature spell you cast each turn costs {2} less to cast and can be cast as though it had flash."
	Register(Spec{
		OracleID:     "84d648c7-655e-4590-a3e0-366a8a4a4cb3",
		Name:         "Radagast of Rhosgobel",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, label, YourSpell(), CreatureSpell(), NoCreatureSpellCastYetThisTurn()),
		},
		CastTimings: []game.CastTimingRule{CastFirstCreatureSpellAsThoughFlash(label)},
	})
}
