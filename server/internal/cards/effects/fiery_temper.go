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
// priced at {R}, keyed "madness" and flash-timed (CR 608.2g).
//
// Declared simplification, shared with cascade and suspend: taking
// the offer grants the cast rather than casting the card inside the
// trigger's resolution, so there is a CR 117.3b response window paper
// does not have. A card accepted and never cast goes to its owner's
// graveyard at the next end step, which keeps the grant from being
// stronger than printed. See game/madness.go.
func init() {
	Register(Spec{
		OracleID:     "f07bd49d-8e71-4d56-be2a-638514011318",
		Name:         "Fiery Temper",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Madness:      "{R}",
		OnResolve:    damageToFirstTarget(3),
	})
}
