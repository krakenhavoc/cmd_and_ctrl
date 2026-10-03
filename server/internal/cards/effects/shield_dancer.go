package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0ca9c400-c62a-49d8-8dce-5caca7f92ed9",
		Name:         "Shield Dancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: The next time target attacking creature would deal combat damage to this creature this turn, that creature deals that damage to itself instead.",
			Cost:    ManaCost("{2}{W}"),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return RedirectDamage{From: t.ID, Protect: ShieldThis, CombatOnly: true, Next: true, To: RedirectToObject(t.ID)}.Apply(ctx)
			},
		}},
	})
}
