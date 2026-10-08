package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Fealty — Sorcery {2}{R}:
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn.
//	 Create a Wicked Role token attached to up to one target creature.
//	 (If you control another Role on it, put that one into the
//	 graveyard. Enchanted creature gets +1/+0. When this token is put
//	 into a graveyard, each opponent loses 1 life.)"
//
// Act of Treason's three clauses, then a second target clause by slot.
// The Role's target is any creature, not only yours or the stolen one,
// and the Role is controlled by the caster either way, so it does not
// replace a Role an opponent controls there (CR 704.5z counts one
// player's Roles).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a5ae75de-0dff-44bf-9dc6-5895ab3ae731",
		Name:         "Twisted Fealty",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetCreature("up to one target creature").WithCount(0, 1),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (GainControl{
					Target:   t.ID,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Twisted Fealty — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Label:    "Twisted Fealty — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return CreateRoleToken{Role: RoleWicked, Host: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
