package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rashida Scalebane — Legendary Creature — Human Soldier {3}{W}{W}, 3/4 (Mirage):
//
//	"{T}: Destroy target attacking or blocking Dragon. It can't be
//	 regenerated. You gain life equal to its power."
//
// The Dragon's power is read before it is destroyed (CR 608.2h), and the life
// is gained whether or not it was actually destroyed ("equal to its power" is
// a separate instruction). The tap cost means summoning sickness applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4aff0f09-0607-41b7-8f60-4ef35ff6183e",
		Name:         "Rashida Scalebane",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target attacking or blocking Dragon. It can't be regenerated. You gain life equal to its power.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking Dragon", effects.AttackingOrBlocking(), effects.Subtype("Dragon")),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				power := card.CurrentPower()
				if err := (effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: power}.Apply(ctx)
			},
		}},
	})
}
