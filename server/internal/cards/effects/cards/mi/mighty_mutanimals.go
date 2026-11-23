package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mighty Mutanimals — {2}{W}{W} Creature — Mutant Rebel 2/1 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"When this creature enters, create a 2/2 red Mutant creature token.
//	 Alliance — Whenever another creature you control enters, put a +1/+1
//	 counter on target creature you control."
//
// Alliance is an ability word (CR 207.2c), Slash, Reptile Rampager's
// trigger: one per other creature that enters under your control, tokens
// included, so the Mutant this creature makes triggers it at once. The
// counter goes on the one target, Agent Bishop's body, and only while the
// target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcdc7dfc-04b5-411a-ab49-282b9c2608e2",
		Name:         "Mighty Mutanimals",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mighty Mutanimals — create a 2/2 red Mutant creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Mutant"), N: 1})),
			effects.Targeting(
				effects.WheneverAnotherCreatureEntersUnderYourControl("Mighty Mutanimals — alliance: put a +1/+1 counter on target creature you control",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
