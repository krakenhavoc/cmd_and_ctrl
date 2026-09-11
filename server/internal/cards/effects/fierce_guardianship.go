package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fierce Guardianship — Instant {2}{U}:
//
//	"If you control a commander, you may cast this spell without
//	paying its mana cost.
//	 Counter target noncreature spell."
//
// The most-played of the Commander Legends free spells: a Negate that
// costs nothing while your commander is out, which is the difference
// between holding up interaction and tapping out.
//
// SANDBOX SIMPLIFICATION — the alternative cost is NOT implemented,
// exactly as on Deadly Rollick. The spell costs {2}{U} here, which is
// strictly WEAKER than printed. See deadly_rollick.go for why the
// engine has no seam for CR 118.9 alternative costs, and
// docs/decklists/top-100-commander-staples.md for the other cards the
// same gap holds back.
//
// The counter half is Negate's, predicate and all: the picker only
// offers noncreature spells and the engine rejects a creature spell
// at announce (CR 601.2c).
func init() {
	Register(Spec{
		OracleID:     "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9",
		Name:         "Fierce Guardianship",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You always pay {2}{U} — the free cast while you control a commander is not offered."},
		Targets:      TargetSpell("target noncreature spell", Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
