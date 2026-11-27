package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last-Ditch Effort — Instant {R}:
//
//	"Sacrifice any number of creatures. Last-Ditch Effort deals that much
//	 damage to any target."
//
// The target is chosen as the spell is cast. As it resolves its controller
// picks which of their creatures to sacrifice (none is allowed), and the
// damage is the number actually sacrificed. With the target illegal the spell
// does nothing, so nothing is sacrificed (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdd8d7cf-28f2-4a4d-a373-40d48bff615b",
		Name:         "Last-Ditch Effort",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			return effects.ChoosePermanents{
				Question:   "Last-Ditch Effort — sacrifice any number of creatures (it deals that much damage to the target)",
				Keyed:      lastDitchEffortSacrificed,
				Candidates: effects.OldB718AnyNumberOfYourCreatures,
			}.Apply(ctx)
		},
	})
}

var lastDitchEffortSacrificed = effects.RegisterPermanentsPick("prompt-run/last-ditch-effort-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	},
})
