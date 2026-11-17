package effects

// Wing Snare — Sorcery {2}{G} (Urza's Legacy, #5623):
//
//	"Destroy target creature with flying."
//
// Plummet at sorcery speed: the pick is a creature with flying, and one that
// has lost flying by resolution is an illegal target, so the spell does
// nothing (CR 608.2b). Regeneration and indestructible apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61ce6ec1-79d8-4dc9-ab2d-1623ed1459c1",
		Name:         "Wing Snare",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve:    destroyTheTargetPermanent,
	})
}
