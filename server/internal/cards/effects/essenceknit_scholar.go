package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essenceknit Scholar — Creature — Dryad Warlock {B}{B/G}{G}, 3/1 (Secrets
// of Strixhaven, #4456):
//
//	"When this creature enters, create a 1/1 black and green Pest creature
//	 token with "Whenever this token attacks, you gain 1 life."
//	 At the beginning of your end step, if a creature died under your
//	 control this turn, draw a card."
//
// The Pest is the Secrets of Strixhaven one (PestAttackToken, life on
// attack, not the Strixhaven Pest's life on death). The end step is
// Faramir's: an intervening "if" (CR 603.4) on the turn's tally of creatures
// that died under your control, tokens included (CR 700.4), checked as the
// step begins and again as the trigger resolves; it triggers once however
// many died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd757328-0593-4fc7-a099-e82f5c0677ff",
		Name:         "Essenceknit Scholar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Essenceknit Scholar — create a 1/1 Pest that gains you life when it attacks",
				Do(CreateToken{Template: PestAttackToken(), N: 1})),
			AtYourEndStepIfACreatureDied("Essenceknit Scholar — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
