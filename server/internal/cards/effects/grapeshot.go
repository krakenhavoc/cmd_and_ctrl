package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grapeshot — Sorcery {1}{R}:
//
//	"Grapeshot deals 1 damage to any target.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn. You may choose new targets for the copies.)"
//
// The storm keyword's canonical card, and the reason it is worth
// having as a declaration rather than a one-off: Grapeshot is
// Lightning Bolt's OnResolve for one damage, plus the word. The whole
// card is four lines because CR 702.40 lives in effects/storm.go and
// game/storm.go, not here (ADR 0086).
//
// Each copy gets its own CR 707.10c "choose new targets" prompt, which
// is the card: a storm count of six is six separate one-damage
// instructions that may point at six different things, not six damage
// to one. The engine's per-copy prompt (spell_copy.go) is what makes
// that true without Grapeshot saying anything about it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebd2d760-5ad8-4027-b124-171822f3edfe",
		Name:         "Grapeshot",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    damageToFirstTarget(1),
	})
}
