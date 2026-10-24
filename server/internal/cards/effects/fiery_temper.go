package effects

// Fiery Temper — {1}{R}{R} Instant:
//
//	"Fiery Temper deals 3 damage to any target.
//	 Madness {R}"
//
// Lightning Bolt with a condition, and the condition is the card: a
// Fiery Temper discarded to a looter costs one red mana instead of
// three, at instant speed, on anybody's turn. It is the madness card
// everyone learns the keyword from, and the one whose Gatherer ruling
// (2022-12-08) says the part the engine had to get right — the
// discard can be a cost, an effect's instruction or the cleanup
// step's hand-size trim, and madness works for all three.
//
// The card file is one damage clause and one string. Both halves of
// the keyword are the engine's (#657): `buildDef` grows the CR 702.35a
// discard replacement and the exile-zone trigger from `Madness`, and
// the cast the trigger offers is a per-instance `game.CastPermission`
// priced at {R}, keyed "madness" and flash-timed (CR 608.2g), and it is
// made inside the trigger's resolution through a resolution_cast prompt
// (#4250): nobody can respond between the discard's trigger and the
// cast, and a decline puts the card into its owner's graveyard then and
// there. See game/madness.go.
func init() {
	Register(Spec{
		OracleID:     "f07bd49d-8e71-4d56-be2a-638514011318",
		Name:         "Fiery Temper",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Madness:      "{R}",
		OnResolve:    damageToFirstTarget(3),
	})
}
