package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Despotic Scepter — Artifact {1} (Ice Age, S82 OLD, #5636):
//
//	"{T}: Destroy target permanent you own. It can't be regenerated."
//
// Staff of Compleation's "you own" target (owner, not controller), without the life payment, and
// the destruction ignores regeneration shields (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34a85d7f-d4ea-4a0f-aa4c-bf0b0f4987bf",
		Name:         "Despotic Scepter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target permanent you own. It can't be regenerated.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target permanent you own", effects.YouOwn()),
			Effect:  effects.S82OldB842DestroyTargetCantRegenerate,
		}},
	})
}
