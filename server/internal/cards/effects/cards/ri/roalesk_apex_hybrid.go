package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roalesk, Apex Hybrid — Legendary Creature — Human Mutant {2}{G}{G}{U},
// 4/5:
//
//	"Flying, trample
//	 When Roalesk enters, put two +1/+1 counters on another target
//	 creature you control.
//	 When Roalesk dies, proliferate, then proliferate again."
//
// The enters trigger targets another creature its controller has (removed
// for want of a target when there is none, CR 603.3d). The dies trigger is
// Ezuri's proliferate twice: each proliferate is its own choice (CR
// 701.34a), the second asked once the first's counters have landed, so a
// counter the first adds can be added to again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca4936cf-48d4-4469-9243-8dbf8ec4cdca",
		Name:            "Roalesk, Apex Hybrid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Roalesk — put two +1/+1 counters on another target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
							return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
						}
						return nil
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
			effects.WhenThisDies("Roalesk — proliferate, then proliferate again",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ProliferateTimesKeyed(g, item.Controller, item.SourceCardID, 2)
				}),
		},
	})
}
