package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ca4936cf-48d4-4469-9243-8dbf8ec4cdca",
		Name:            "Roalesk, Apex Hybrid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Roalesk — put two +1/+1 counters on another target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if id, ok := b16FirstLegalTargetCard(ctx); ok {
							return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
						}
						return nil
					}),
				Another(TargetCreature("another target creature you control", YouControl()))),
			WhenThisDies("Roalesk — proliferate, then proliferate again",
				func(g *game.Game, item *game.StackItem) error {
					return proliferateTimesKeyed(g, item.Controller, item.SourceCardID, 2)
				}),
		},
	})
}
