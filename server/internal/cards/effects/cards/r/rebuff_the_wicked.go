package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rebuff the Wicked — Instant {W} (EDHREC rank 1547):
//
//	"Counter target spell that targets a permanent you control."
//
// White's one-mana counterspell, narrow by design: the clause is a
// predicate over the targeted spell's own targets, read off its stack
// item — at least one of them must be a permanent the caster controls
// ON THE BATTLEFIELD right now. A spell aimed at the caster's face
// does not qualify (a player is not a permanent), and neither does
// one whose only permanent target has already left, because the
// predicate re-runs at resolution like every target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5822e53-ddec-4c77-bcf4-091e35c1e731",
		Name:         "Rebuff the Wicked",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell that targets a permanent you control", effects.B14SpellTargetsAPermanentYouControl()),
		OnResolve:    effects.SomB1CounterTheTargetSpells,
	})
}
