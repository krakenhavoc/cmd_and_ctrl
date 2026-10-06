package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Al-abara's Carpet — Artifact {5}:
//
//	"{5}, {T}: Prevent all damage that would be dealt to you this turn by
//	 attacking creatures without flying."
//
// #2026's combat status and keyword negation, read as each creature
// would deal damage (CR 609.7b): combat damage or not, to you only. A
// creature that has left the battlefield is read as it last existed
// there (CR 608.2h): its damage is prevented if it was an attacking
// creature without flying when it left (Heavy Fog's ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0f5b0c77-1e3d-46a1-ae0e-03ed79196cd9",
		Name:         "Al-abara's Carpet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{sourceShieldRow(
			"{5}, {T}: Prevent all damage that would be dealt to you this turn by attacking creatures without flying.",
			Plus(ManaCost("{5}"), TapCost()), nil,
			PreventDamageFromSource{Protect: ShieldYou, Queries: creatureSources(), Filter: game.DamageSourceFilter{
				Combat: game.SourceCombatAttacking,
				Except: []game.PermanentQuery{{Keyword: "flying"}},
			}})},
	})
}
