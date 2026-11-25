package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Domri, City Smasher — Legendary Planeswalker — Domri {4}{R}{G} (Ravnica
// Allegiance), loyalty 4:
//
//	"+2: Creatures you control get +1/+1 and gain haste until end of turn.
//	 −3: Domri deals 3 damage to any target.
//	 −8: Put three +1/+1 counters on each creature you control. Those creatures
//	 gain trample until end of turn."
//
// The +2 and the −8 act on the creatures you control as the ability resolves
// (CR 611.2c): one that enters afterwards gets nothing. The −8 puts the
// counters first and grants trample to the same set.
//
// No simplification.
func init() {
	mine := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:        "d8fe9672-0ead-4f47-abd8-e1ee133c5c60",
		Name:            "Domri, City Smasher",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Creatures you control get +1/+1 and gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(2),
				Effect: effects.Do(
					effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Domri, City Smasher — +1/+1"},
					effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"haste"}, Label: "Domri, City Smasher — haste"},
				),
			},
			{
				Label:   "−3: Domri deals 3 damage to any target.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
			},
			{
				Label: "−8: Put three +1/+1 counters on each creature you control. Those creatures gain trample until end of turn.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.PutPlusOneCountersOnEachYouControlThat(3, effects.Creature())(g, item); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Domri, City Smasher — trample"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
