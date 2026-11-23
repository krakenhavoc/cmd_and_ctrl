package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Lectern — Artifact Creature — Construct {1}{U}, 0/4:
//
//	"{1}, Sacrifice this creature: Draw a card. Create a Sorcerer Role
//	 token attached to up to one other target creature you control.
//	 Activate only as a sorcery. (If you control another Role on it, put
//	 that one into the graveyard. Enchanted creature gets +1/+1 and has
//	 "Whenever this creature attacks, scry 1.")"
//
// The Role goes on the target only if it is still a creature on the
// battlefield when the ability resolves (CR 608.2b); the draw happens
// either way. The sacrifice is a cost, so the Lectern is gone before
// either instruction, and "other" keeps it from naming itself.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "061e7fb2-2c04-4845-b60a-0e2512a2dec1",
		Name:         "Living Lectern",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, Sacrifice this creature: Draw a card. Create a Sorcerer Role token attached to up to one other target creature you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets:      effects.Another(effects.TargetCreature("up to one other target creature you control", effects.YouControl()).WithCount(0, 1)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(effects.NewContext(g, item)); err != nil {
					return err
				}
				return effects.CreateRoleOnFirstTarget(effects.RoleSorcerer)(g, item)
			},
		}},
	})
}
