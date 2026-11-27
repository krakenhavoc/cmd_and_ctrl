package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raka Disciple — Creature — Minotaur Wizard {R}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"{W}, {T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn.
//	 {U}, {T}: Target creature gains flying until end of turn."
//
// Two tap abilities (CR 302.6 applies to both): Jhessian Balmgiver's one-damage
// shield on any target, and a flying grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88c5a681-4095-4e37-9149-002bcbaf4c0b",
		Name:         "Raka Disciple",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.OldB632PreventOneRow("Raka Disciple", "{W}"),
			effects.OldB630DiscipleFlyingRow("Raka Disciple", "{U}"),
		},
	})
}
