package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloody Betrayal — Sorcery {2}{R} (Innistrad: Crimson Vow):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. Create a Blood token."
//
// Act of Treason's three clauses, then the Blood token. If the target is
// illegal on resolution the whole spell does nothing (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "a2743cb6-564d-43f1-99cc-4ba3ceb2e282",
		Name:         "Bloody Betrayal",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Bloody Betrayal — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Bloody Betrayal — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return s82Vow03CreateBlood(1).Apply(ctx)
		},
	})
}
