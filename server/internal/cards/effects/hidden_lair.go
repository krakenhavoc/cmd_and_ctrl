package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Lair — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {T}: Add {U} or {B}. Activate only if this land entered this turn or
//	 if you control a basic land."
//
// Two mana abilities (CR 605.1a). The second carries the printed gate as
// its Condition, read when the ability is activated: the land's own entry
// is the turn tally (game.EnteredThisTurn, CR 400.7), not summoning
// sickness, and "a basic land" is the Basic supertype on a permanent you
// control (Hidden Lair itself is not basic).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7069d241-4e66-40bf-afd1-551a4a5457f0",
		Name:         "Hidden Lair",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			mshA4GatedDual("U", "B"),
		},
	})
}

// mshA4GatedDual is a "{T}: Add {A} or {B}. Activate only if this land
// entered this turn or if you control a basic land" row.
func mshA4GatedDual(a, b string) ManaAbility {
	row := dualManaAbility(a, b)
	row.Label += ". Activate only if this land entered this turn or if you control a basic land."
	basic := ControlsAtLeast(1, IsBasicLand)
	row.Condition = func(g *game.Game, controller, source uuid.UUID) bool {
		return g.EnteredThisTurn(source) || basic(g, controller, source)
	}
	return row
}
