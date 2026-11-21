package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voltaic Key — Artifact {1} (EDHREC rank 1752):
//
//	"{1}, {T}: Untap target artifact."
//
// The one-mana untapper. A mana-plus-tap activated ability with an
// artifact target clause; the untap is the primitive. "Target
// artifact" admits the Key itself, as printed — pointless, but the
// rules allow it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09aeea91-b1dc-443f-a509-4758f052c0a7",
		Name:         "Voltaic Key",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Untap target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
