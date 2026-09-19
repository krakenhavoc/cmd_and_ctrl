package effects

// Spellbook — Artifact — Book {0}:
//
//	"You have no maximum hand size."
//
// A free artifact whose whole text is one player-level continuous
// effect, and the engine already carries it: Spec.NoMaxHandSize is
// the derived static Reliquary Tower, Thought Vessel and Decanter of
// Endless Water share. The cleanup step asks the battlefield
// (game.EffectiveMaxHandSizeLocked) rather than writing to
// Player.MaxHandSize, so a second copy entering and the first one
// leaving both come out right with no bookkeeping of their own.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "83133f51-bfbd-4db5-9be0-f660e3b66435",
		Name:          "Spellbook",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
	})
}
