package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Donna Noble — Legendary Creature — Human {3}{R}, 2/4 (Doctor Who):
//
//	"Soulbond (You may pair this creature with another unpaired creature when
//	 either enters. They remain paired for as long as you control both of them.)
//	 Whenever Donna Noble or a creature it's paired with is dealt damage, Donna
//	 Noble deals that much damage to target opponent.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// Soulbond is the engine's pairing (Soulbond, CR 702.95). The trigger fires once per
// damage event Donna or her current partner is dealt (the partner is read as the
// damage is dealt, CR 702.95), chooses its target opponent as it goes on the stack,
// and deals Donna's damage equal to the amount of that event (last-known information
// if she has since left, CR 608.2h; damage prevented was never dealt). Doctor's
// companion is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0d6e76aa-2966-48b7-be4e-2d3e2cc94fc1",
		Name:            "Donna Noble",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered: append(Soulbond("Donna Noble"),
			Targeting(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Amount <= 0 {
						return false
					}
					if ev.Target == source.InstanceID {
						return true
					}
					return ev.Target != uuid.Nil && g.PairedWithForEffect(source.InstanceID) == ev.Target
				}, "Donna Noble — she deals that much damage to target opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						ts := ctx.LegalTargets()
						n := ctx.Trigger().Event.Amount
						if len(ts) == 0 || n <= 0 {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: n}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent()))),
	})
}
