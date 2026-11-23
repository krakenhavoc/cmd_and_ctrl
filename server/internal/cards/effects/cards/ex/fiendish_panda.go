package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b5d64271-4478-4c55-8c71-ba8f02b47fac",
		Name:         "Fiendish Panda",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Fiendish Panda — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.Targeting(
				effects.WhenThisDies("Fiendish Panda — return another target non-Bear creature card with mana value less than or equal to its power from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.RelativeToSource(
					effects.Another(effects.TargetCardInGraveyard("another target non-Bear creature card with mana value less than or equal to Fiendish Panda's power from your graveyard",
						effects.Creature(), effects.YouOwn(), effects.Not(effects.HasSubtype("Bear")))),
					effects.ManaValueNoGreaterThanPower())),
		},
	})
}
