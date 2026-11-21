package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn the Tables — Instant {3}{W}{W}:
//
//	"All combat damage that would be dealt to you this turn is dealt to
//	 target attacking creature instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn of combat
// damage to you, to the target. The rulings: a target that has left
// redirects nothing; one removed from combat but still on the
// battlefield is still dealt the damage.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd5ad7c3-4477-4278-9875-d2ffbb0e8089",
		Name:         "Turn the Tables",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.RedirectSpell(effects.RedirectDamage{Protect: effects.ShieldYou, CombatOnly: true, To: effects.RedirectToClause(0)}),
	})
}
