package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarecrow — Artifact Creature — Scarecrow {5}, 2/2:
//
//	"{6}, {T}: Prevent all damage that would be dealt to you this turn by creatures with flying."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record with no
// source and the property "a creature with flying", checked as the damage
// would be dealt (CR 615.9), protecting you. Non-combat damage from a
// flying creature is prevented too (its ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "58616e15-531f-4680-8294-35ab7e962c96",
		Name:         "Scarecrow",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"{6}, {T}: Prevent all damage that would be dealt to you this turn by creatures with flying.",
			Plus(ManaCost("{6}"), TapCost()), nil,
			PreventDamageFromSource{Protect: ShieldYou, Queries: []game.PermanentQuery{{Types: []string{"creature"}, Keyword: "flying"}}})},
	})
}
