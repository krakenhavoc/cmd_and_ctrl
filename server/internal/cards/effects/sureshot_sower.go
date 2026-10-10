package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5bbc9d7d-91bd-4a58-8d19-f7be650720b7",
		Name:            "Sureshot Sower",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, Discard this card: Destroy target creature with flying.",
			Cost:    game.AbilityCost{Mana: "{3}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return DestroyTarget{Target: ts[0].ID}.Apply(ctx)
			},
		}},
	})
}
