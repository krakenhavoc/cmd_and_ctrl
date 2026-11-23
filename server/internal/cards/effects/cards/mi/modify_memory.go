package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Modify Memory — Sorcery {4}{U} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Exchange control of two target creatures controlled by different players. If
//	 you control neither creature, draw three cards."
//
// Visions of Duplicity's exchange (CR 701.12, all or nothing: nothing changes
// hands if either target is illegal as the spell resolves) over two creatures
// whose controllers differ (a set rule on the one clause, Molten Primordial's).
// "If you control neither" is read from the targeted creatures' controllers after
// the exchange: a creature the caster controlled before is theirs or the other's
// either way, so the answer is the same on both readings of the clause, and a
// failed exchange (a target gone) still pays out the three cards when the
// caster controls neither.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c37d4e1-6b76-4feb-933b-1752ecc6d3e8",
		Name:         "Modify Memory",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("two target creatures controlled by different players").
			WithCount(2, 2).EachDifferent(effects.EachDifferentController()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 2 {
				if _, err := (effects.ExchangeControl{A: legal[0].ID, B: legal[1].ID, Label: "Modify Memory — exchange control"}).ApplyAndReport(ctx); err != nil {
					return err
				}
			}
			for _, t := range item.Targets {
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.Controller == item.Controller {
					return nil
				}
			}
			return effects.DrawCards{N: 3}.Apply(ctx)
		},
	})
}
