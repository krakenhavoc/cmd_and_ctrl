package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keening Stone — Artifact {6}:
//
//	"{5}, {T}: Target player mills X cards, where X is the number of cards in
//	 that player's graveyard."
//
// X is counted as the ability resolves, before any card is milled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0f06250-25a5-4003-8f3a-6d85ef32ea10",
		Name:         "Keening Stone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Target player mills X cards, where X is the number of cards in that player's graveyard",
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					p := g.PlayerByIDForEffect(t.ID)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					return MillCards{Player: t.ID, N: len(p.Graveyard.Cards)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
