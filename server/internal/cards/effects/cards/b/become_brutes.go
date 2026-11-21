package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Become Brutes — Sorcery {1}{R}:
//
//	"One or two target creatures each gain haste until end of turn. For
//	 each of those creatures, create a Monster Role token attached to
//	 it. (If you control another Role on it, put that one into the
//	 graveyard. Enchanted creature gets +1/+1 and has trample.)"
//
// "Those creatures" are the targets still legal at resolution
// (CR 608.2b). The haste is granted to all of them first, as printed,
// then one Role each. The Roles are the caster's, whoever controls the
// creature.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "789e034c-7f4a-4fd5-9dd8-48313742126e",
		Name:         "Become Brutes",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			for _, t := range targets {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Label:    "Become Brutes — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range targets {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.CreateRoleToken{Role: effects.RoleMonster, Host: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
