package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snuff Out — Instant {3}{B}:
//
//	"If you control a Swamp, you may pay 4 life rather than pay this
//	 spell's mana cost.
//	 Destroy target nonblack creature. It can't be regenerated."
//
// The issue listed this under "additional cost", which it is not:
// "RATHER THAN pay this spell's mana cost" is CR 118.9's alternative
// cost, the same clause overload and evoke ride. An additional cost
// would be paid on top.
//
// The Swamp check gates the OFFER, so a player with no Swamp is never
// shown the free option — and the engine refuses the claim if the
// Swamp leaves in the window between snapshot and announce.
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c):
// the rider rides the destroy route onto the CR 614 event and the
// regeneration built-in declines. The shield is not spent
// (CR 701.19c).
//
// No simplification: reviewed for the Mass Casualty deck request
// (#3245).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "324824cb-f938-401c-b9b5-d8908b431ef0",
		Name:         "Snuff Out",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		AlternativeCosts: []game.AlternativeCost{
			effects.PayLifeInstead("Pay 4 life (you control a Swamp)", 4, effects.ControlsA("Swamp")),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID, CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
