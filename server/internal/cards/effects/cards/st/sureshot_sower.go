package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sureshot Sower — Creature — Human Archer {1}{G}, 3/1:
//
//	"Reach
//	 {3}{G}, Discard this card: Destroy target creature with flying."
//
// The hand ability (Boseiju's channel shape): it is activated from the
// hand, with the card itself discarded as part of the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5bbc9d7d-91bd-4a58-8d19-f7be650720b7",
		Name:            "Sureshot Sower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, Discard this card: Destroy target creature with flying.",
			Cost:    game.AbilityCost{Mana: "{3}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.DestroyTarget{Target: ts[0].ID}.Apply(ctx)
			},
		}},
	})
}
