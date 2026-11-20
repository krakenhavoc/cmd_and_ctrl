package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Efflorescence — Instant {2}{G} (Secrets of Strixhaven, #4456):
//
//	"Put two +1/+1 counters on target creature.
//	 Infusion — If you gained life this turn, that creature also gains
//	 trample and indestructible until end of turn."
//
// Infusion is an ability word (CR 207.2c) and does nothing itself. The
// counters always go on; "if you gained life this turn" is read as the
// spell resolves, off the caster's per-turn tally of life gained, and the
// same creature then gains both keywords until end of turn. A target that
// became illegal takes neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "181da74b-53fb-4e46-afc0-e7c990fda128",
		Name:         "Efflorescence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample", "indestructible")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				if effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller()) == 0 {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"trample", "indestructible"},
					Label:    "Efflorescence — trample and indestructible until end of turn",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
