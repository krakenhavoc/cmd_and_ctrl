package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Destined White Mage — Legendary Creature — Human Cleric {2}{W}, 2/3
// (Final Fantasy, #4462):
//
//	"Lifelink
//	 {W}, {T}: Another target creature you control gains lifelink until
//	 end of turn.
//	 Whenever you gain life, put a +1/+1 counter on target creature you
//	 control. If you have a full party, put three +1/+1 counters on that
//	 creature instead."
//
// Every life-gain event is its own trigger (Cradle of Vitality's
// WheneverYouGainLife), with the creature chosen as it goes on the stack.
// The party is counted as the trigger resolves (CR 700.8c), so the three
// counters follow the board then; a target that left does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8110200-209e-423a-96b7-81a3ac32447c",
		Name:            "The Destined White Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouGainLife("The Destined White Mage — put a +1/+1 counter on target creature you control, or three with a full party",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						target, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						n := 1
						if finB9HasFullParty(g, item.Controller) {
							n = 3
						}
						return AddCounter{Target: target, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Another target creature you control gains lifelink until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  ebTargetCreatureGainsUntilEOT("lifelink", "The Destined White Mage — lifelink until end of turn"),
		}},
	})
}
