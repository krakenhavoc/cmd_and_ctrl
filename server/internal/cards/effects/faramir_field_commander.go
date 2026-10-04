package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faramir, Field Commander — Legendary Creature — Human Soldier
// {3}{W}, 3/3:
//
//	"At the beginning of your end step, if a creature died under your
//	 control this turn, draw a card.
//	 Whenever the Ring tempts you, if you chose a creature other than
//	 Faramir as your Ring-bearer, create a 1/1 white Human Soldier
//	 creature token."
//
// The end-step "if" is checked as the step begins and again as the
// ability resolves (CR 603.4); Faramir need not have been on the
// battlefield when the creature died (2023-06-16 ruling). The second
// "if" is a fact about the temptation that triggered it
// (IfYouChoseAnotherRingBearer): choosing no creature, or Faramir,
// makes no token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adeed5e8-9019-4740-83ea-67dcd754e194",
		Name:         "Faramir, Field Commander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfACreatureDied("Faramir, Field Commander — draw a card", Do(DrawCards{N: 1})),
			IfYouChoseAnotherRingBearer(WheneverTheRingTemptsYou("Faramir, Field Commander — create a 1/1 white Human Soldier",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1}))),
		},
	})
}
