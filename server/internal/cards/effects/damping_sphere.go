package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damping Sphere — Artifact {2}:
//
//	"If a land is tapped for two or more mana, it produces {C}
//	 instead of any other type and amount.
//	 Each spell a player casts costs {1} more to cast for each other
//	 spell that player has cast this turn."
//
// The storm half is the one implemented here, and it is the
// catalog's first SCALING cost modifier: the amount is a function of
// the board rather than a printed number, which is the whole reason
// CostModifier.Amount is a hook.
//
// "Each OTHER spell that player has cast this turn" comes out right
// for free. The per-turn tally (Game.SpellsCastThisTurn) is bumped
// AFTER the spell reaches the stack, and CR 601.2f prices the cast
// before that, so the count this reads already excludes the spell
// being priced. The first spell of a turn is taxed nothing, the
// second {1}, the third {2}.
//
// SANDBOX SIMPLIFICATION — the Ancient Tomb / Cabal Coffers half
// ("if a land is tapped for two or more mana, it produces {C}
// instead") is NOT implemented. It is a replacement effect on mana
// production, and the mana pipeline (#352) has no seam for
// replacing an ability's output between ProducedFunc and the pool.
// Strictly weaker than printed: the Sphere taxes storm here and
// leaves the big-mana lands alone.
func init() {
	Register(Spec{
		OracleID:     "fd50d76c-7654-47b9-a5b6-d075874e4357",
		Name:         "Damping Sphere",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only the storm tax works — a land tapped for two or more mana still produces what it normally would, not {C}."},
		CostModifiers: []game.CostModifier{
			CostsMoreEach(OtherSpellsCastThisTurn(),
				"Each spell a player casts costs {1} more to cast for each other spell that player has cast this turn."),
		},
	})
}
