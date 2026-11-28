package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ana Disciple — Creature — Human Wizard {G}, 1/1 (Apocalypse, S82 OLD, #5614):
//
//	"{U}, {T}: Target creature gains flying until end of turn.
//	 {B}, {T}: Target creature gets -2/-0 until end of turn."
//
// Two targeted tap abilities (CR 302.6 applies to both).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f543dfcd-015e-48bc-851d-08002d0241fa",
		Name:         "Ana Disciple",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.OldB630DiscipleFlyingRow("Ana Disciple", "{U}"),
			effects.OldB630DiscipleBoostRow("Ana Disciple", "{B}", -2),
		},
	})
}
