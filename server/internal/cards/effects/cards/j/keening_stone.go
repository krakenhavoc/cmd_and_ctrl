package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keening Stone — Artifact {6}:
//
//	"{5}, {T}: Target player mills X cards, where X is the number of cards in
//	 that player's graveyard."
//
// X is counted as the ability resolves, before any card is milled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0f06250-25a5-4003-8f3a-6d85ef32ea10",
		Name:         "Keening Stone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Target player mills X cards, where X is the number of cards in that player's graveyard",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					p := g.PlayerByIDForEffect(t.ID)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					return effects.MillCards{Player: t.ID, N: len(p.Graveyard.Cards)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
