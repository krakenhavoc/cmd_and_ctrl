package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrorform — Instant {4}{U}{U} (Lorwyn Eclipsed, #4458):
//
//	"Each nonland permanent you control becomes a copy of target
//	 non-Aura permanent."
//
// Mirrorweave's one copy over many permanents (become_copy.go), with no
// duration (CR 611.2a): the affected set is every nonland permanent the
// caster controls as the spell resolves (CR 611.2c), and the copiable
// values are read once, from the target, as the effect begins (CR 707.2),
// so a target that is itself a copy hands on what it copies and its
// counters, tapped state and non-copy effects are not copied (the
// rulings). Nothing enters, so no enters ability of the copied permanent
// applies. The target, when it is one of the caster's own, becomes a copy
// of itself, which changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81cb888c-c4d3-45a2-a03e-bfd91a2af2be",
		Name:         "Mirrorform",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target non-Aura permanent", effects.Not(effects.Subtype("Aura"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			var mine []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && !c.IsLand() {
					mine = append(mine, c.InstanceID)
				}
			}
			return effects.BecomeCopy{
				Targets:  mine,
				Of:       targets[0].ID,
				Duration: effects.CopyIndefinite,
				Label:    "Mirrorform — becomes a copy",
			}.Apply(ctx)
		},
	})
}
