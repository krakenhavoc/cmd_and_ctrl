package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metallic Mastery — Sorcery {2}{R} (Mirrodin Besieged, S82 OLD b465, #5576):
//
//	"Gain control of target artifact until end of turn. Untap that artifact.
//	 It gains haste until end of turn."
//
// Act of Treason on an artifact: a layer-2 control effect that lapses at
// cleanup, the untap, and a layer-6 haste grant (a stolen permanent is
// summoning-sick under its new controller, CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7de034a4-a42a-4e60-9ec9-5e9bb9987cfa",
		Name:         "Metallic Mastery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GainControl{
				Target:   t.ID,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Metallic Mastery — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"haste"},
				Label:    "Metallic Mastery — haste until end of turn",
			}.Apply(ctx)
		},
	})
}
