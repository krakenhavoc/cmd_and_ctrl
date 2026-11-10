package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiendish Panda — Creature — Bear Demon {2}{W}{B}, 3/2 (Foundations,
// #4465):
//
//	"Whenever you gain life, put a +1/+1 counter on this creature.
//	 When this creature dies, return another target non-Bear creature card
//	 with mana value less than or equal to this creature's power from your
//	 graveyard to the battlefield."
//
// Aerith Gainsborough's counter trigger, and Julian Bashir's bounded
// return: the mana-value cap is data on the clause (RelativeToSource with
// ManaValueNoGreaterThanPower), judged against the Panda's power as the
// target is chosen and again as the ability resolves (CR 608.2b), and
// against its last-known power, counters included, once it has left (the
// printed ruling, CR 608.2h). The Panda is in the graveyard by then and
// "another" keeps it from returning itself; a Bear card, the Panda's own
// type included, is not a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5d64271-4478-4c55-8c71-ba8f02b47fac",
		Name:         "Fiendish Panda",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Fiendish Panda — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			Targeting(
				WhenThisDies("Fiendish Panda — return another target non-Bear creature card with mana value less than or equal to its power from your graveyard to the battlefield",
					returnFirstLegalGraveyardTargetToBattlefield),
				RelativeToSource(
					Another(TargetCardInGraveyard("another target non-Bear creature card with mana value less than or equal to Fiendish Panda's power from your graveyard",
						Creature(), YouOwn(), Not(HasSubtype("Bear")))),
					ManaValueNoGreaterThanPower())),
		},
	})
}
