package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield Dancer — Creature — Human Rebel {2}{W}, 1/3:
//
//	"{2}{W}: The next time target attacking creature would deal combat
//	 damage to this creature this turn, that creature deals that damage
//	 to itself instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection of the target's combat
// damage to the Dancer, dealt to the target itself.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ca9c400-c62a-49d8-8dce-5caca7f92ed9",
		Name:         "Shield Dancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: The next time target attacking creature would deal combat damage to this creature this turn, that creature deals that damage to itself instead.",
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.RedirectDamage{From: t.ID, Protect: effects.ShieldThis, CombatOnly: true, Next: true, To: effects.RedirectToObject(t.ID)}.Apply(ctx)
			},
		}},
	})
}
