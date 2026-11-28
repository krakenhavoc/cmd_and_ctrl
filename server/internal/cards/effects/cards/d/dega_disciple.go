package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dega Disciple — Creature — Human Wizard {W}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"{B}, {T}: Target creature gets -2/-0 until end of turn.
//	 {R}, {T}: Target creature gets +2/+0 until end of turn."
//
// Two targeted tap abilities (CR 302.6 applies to both).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6c53277-daad-4aa3-b29a-11378d550651",
		Name:         "Dega Disciple",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.OldB630DiscipleBoostRow("Dega Disciple", "{B}", -2),
			effects.OldB630DiscipleBoostRow("Dega Disciple", "{R}", 2),
		},
	})
}
