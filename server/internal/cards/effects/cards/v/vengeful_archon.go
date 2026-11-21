package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeful Archon — Creature — Archon {4}{W}{W}{W}, 7/7:
//
//	"Flying
//	 {X}: Prevent the next X damage that would be dealt to you this turn. If damage is prevented this way, this creature deals that much damage to target player or planeswalker."
//
// ADR 0108 owner decision 2 (#1906): a charged shield on you (CR 615.7)
// whose CR 615.5 additional effect has the Archon deal the amount prevented
// to the ability's target, carried on the shield as its follow-up's
// recipient (game.ShieldFollowUp To). Each activation is its own shield,
// and the rulings hold: the next X damage from any source, whenever it
// comes; an illegal target at resolution means no shield; after
// resolution the target's legality is not checked again, only whether it
// is still there; the damage is the Archon's own, not combat damage, and
// nothing is dealt for damage that can't be prevented (CR 615.12).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3ef4311-c7f1-45ae-8bd5-9c05bdf2ae88",
		Name:            "Vengeful Archon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}: Prevent the next X damage that would be dealt to you this turn. If damage is prevented this way, this creature deals that much damage to target player or planeswalker.",
			Cost:    effects.ManaCost("{X}"),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				to, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.PreventNextDamage{
					Target: ctx.Controller(),
					Amount: ctx.X(),
					Then:   effects.DealThatMuchToTheChosenTargetBody,
					To:     to.ID,
					Label:  "Vengeful Archon — prevent the next damage to you and deal that much",
				}.Apply(ctx)
			},
		}},
	})
}
