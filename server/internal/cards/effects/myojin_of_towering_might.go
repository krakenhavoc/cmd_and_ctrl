package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myojin of Towering Might — Legendary Creature — Spirit {5}{G}{G}{G},
// 8/8 (Kamigawa: Neon Dynasty Commander, #5493):
//
//	"Myojin of Towering Might enters with an indestructible counter on it
//	 if you cast it from your hand.
//	 Remove an indestructible counter from Myojin of Towering Might:
//	 Distribute eight +1/+1 counters among any number of target creatures
//	 you control. They gain trample until end of turn."
//
// The counter is Myojin of Night's Reach's entry clause; the division is
// Blessings of Nature's (CR 601.2d, announced with the targets) and the
// trample is Ajani, Sleeper Agent's, granted to each target still legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "1cb42cf5-ae69-46c5-b2ff-d2c1cd60099b",
		Name:                       "Myojin of Towering Might",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: neoMyojinEntry(),
		Activated: []ActivatedAbility{{
			Label: "Remove an indestructible counter from Myojin of Towering Might: Distribute eight +1/+1 counters among any number of target creatures you control. They gain trample until end of turn.",
			Cost:  neoRemoveIndestructible(),
			// ADR 0142: counters and trample answer a removal spell and a block.
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Targets: TargetCreature("any number of target creatures you control", YouControl()).
				WithCount(0, 0).Dividing(Divide(8)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"},
						Label: "Myojin of Towering Might — trample until end of turn"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
