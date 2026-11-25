package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hijack — Sorcery {1}{R}{R} (Ixalan, S82 older sets):
//
//	"Gain control of target artifact or creature until end of turn. Untap it.
//	 It gains haste until end of turn."
//
// Systems Override's three steps (a control change until end of turn, an
// untap, haste until end of turn) without its Spacecraft clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd5651f7-52f8-4845-8a1b-32e5d2e9b00e",
		Name:         "Hijack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{Target: t.ID, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Hijack — gain control until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Hijack — haste until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
