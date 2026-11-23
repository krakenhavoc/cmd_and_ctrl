package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimian Night Stalker — Creature — Nightstalker {3}{B}{B}, 4/4:
//
//	"{B}, {T}: All damage that would be dealt to you this turn by target
//	 attacking creature is dealt to this creature instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn of the
// target's damage to you, combat or not (the ruling), to the Night
// Stalker as it is now.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09b9e6fd-7a61-4ed4-a121-61b64fbf03f4",
		Name:         "Shimian Night Stalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: All damage that would be dealt to you this turn by target attacking creature is dealt to this creature instead.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.RedirectDamage{From: t.ID, Protect: effects.ShieldYou, To: effects.RedirectToThis}.Apply(ctx)
			},
		}},
	})
}
