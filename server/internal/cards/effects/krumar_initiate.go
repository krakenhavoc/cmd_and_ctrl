package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krumar Initiate — Creature — Human Cleric {1}{B}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"{X}{B}, {T}, Pay X life: This creature endures X. Activate only as a
//	 sorcery. (Put X +1/+1 counters on it or create an X/X white Spirit
//	 creature token.)"
//
// "Pay X life" is the cost component AbilityCost.LifeX (PayXLife): the X
// the mana's {X} announces is also the life paid (CR 107.3a), so the life
// total is a ceiling on X (CR 119.4) in the engine's announce gate and in
// the legal-move enumerator. The endure is EndureN with the announced X;
// an Initiate that has left by resolution can't take counters, so the
// Spirit is created (CR 701.63a). X = 0 pays nothing and endures nothing
// (CR 701.63b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65a7fb13-77b9-44ea-9e0e-34daef82eb69",
		Name:         "Krumar Initiate",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:        "{X}{B}, {T}, Pay X life: This creature endures X. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{X}{B}"), TapCost(), PayXLife()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return EndureN(ctx, item.SourceCardID, ctx.X(), "Krumar Initiate")
			},
		}},
	})
}
