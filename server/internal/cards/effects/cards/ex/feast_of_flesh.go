package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast of Flesh — Sorcery {B} (Coldsnap):
//
//	"Feast of Flesh deals X damage to target creature and you gain X life,
//	 where X is 1 plus the number of cards named Feast of Flesh in all
//	 graveyards."
//
// X is counted once as the spell resolves, while Feast of Flesh is still on
// the stack and so not itself in a graveyard (CR 608.2h, CR 608.2n). The life
// is gained whether or not the damage is prevented. With no legal target the
// spell is countered on resolution and gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f214306-1664-41ec-b999-1b7471bbcbd6",
		Name:         "Feast of Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := 1
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Feast of Flesh" {
						x++
					}
				}
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: item.Controller, Amount: x}.Apply(ctx)
		},
	})
}
