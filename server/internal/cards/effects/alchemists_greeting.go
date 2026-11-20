package effects

// Alchemist's Greeting — Sorcery {4}{R} (Eldritch Moon, #5541):
//
//	"Alchemist's Greeting deals 4 damage to target creature.
//	 Madness {1}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Four damage to one creature (damageToFirstTarget) and the one Madness
// string, which grows the discard replacement and the exile trigger
// (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9aaf0df7-5e2d-4f49-a8cb-66523be15ad6",
		Name:         "Alchemist's Greeting",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Madness:      "{1}{R}",
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageToFirstTarget(4),
	})
}
