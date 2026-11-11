package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d96763a0-6a6e-4520-899a-468b4bb307c8",
		Name:         "Mass Mutiny",
		Completeness: CompletenessFull,
		Targets:      perOpponentClauses(game.MaxPlayers-1, "creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{
					Target:   t.ID,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Mass Mutiny — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
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
