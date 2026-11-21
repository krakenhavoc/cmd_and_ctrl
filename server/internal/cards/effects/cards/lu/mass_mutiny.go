package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mass Mutiny — Sorcery {3}{R}{R}:
//
//	"For each opponent, gain control of up to one target creature that
//	 player controls until end of turn. Untap those creatures. They gain
//	 haste until end of turn."
//
// "For each opponent, up to one target … that player controls" is
// perOpponentClauses: one "up to one" clause per opponent, each bound to
// that opponent, so the CR 608.2b re-check asks whether THAT player still
// controls the pick. Each pick still legal is stolen, untapped and given
// haste, in printed order, exactly as Act of Treason does it for one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d96763a0-6a6e-4520-899a-468b4bb307c8",
		Name:         "Mass Mutiny",
		Completeness: effects.CompletenessFull,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GainControl{
					Target:   t.ID,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Mass Mutiny — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Label:    "Mass Mutiny — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
