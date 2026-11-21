package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "c8110200-209e-423a-96b7-81a3ac32447c",
		Name:            "The Destined White Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouGainLife("The Destined White Mage — put a +1/+1 counter on target creature you control, or three with a full party",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						target, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						n := 1
						if effects.FinB9HasFullParty(g, item.Controller) {
							n = 3
						}
						return effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Another target creature you control gains lifelink until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("lifelink", "The Destined White Mage — lifelink until end of turn"),
		}},
	})
}
