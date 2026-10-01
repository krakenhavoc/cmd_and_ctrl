package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nine Lives — Enchantment {1}{W}{W}:
//
//	"Hexproof
//	 If a source would deal damage to you, prevent that damage and put an
//	 incarnation counter on this enchantment.
//	 When there are nine or more incarnation counters on this
//	 enchantment, exile it.
//	 When this enchantment leaves the battlefield, you lose the game."
//
// ADR 0107 §1 (#1858). Hexproof is the printed keyword. The second line is
// a prevention effect (CR 615.1a) with an additional effect (CR 615.5):
// each damage event to the controller is prevented whole and puts ONE
// counter on, however much it was. The exile is a CR 603.8 state trigger,
// and whatever takes the enchantment off the battlefield — that exile, a
// Disenchant, a bounce — triggers the loss (CR 603.6c, 104.3e). "You" is
// its controller as it left.
//
// No simplification.
func init() {
	const incarnation = "incarnation"
	Register(Spec{
		OracleID:        "236e1f57-7ef5-455a-82a6-8ff6b85d8849",
		Name:            "Nine Lives",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Replacements: []game.ReplacementEffect{
			damageToYouBecomesCounters(incarnation, false, "Nine Lives — prevent that damage and put an incarnation counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisHasAtLeast(incarnation, 9, "Nine Lives — exile it", ExileThisThen(nil)),
			On(game.EventLTB, Self, "Nine Lives — you lose the game", Do(LoseTheGame{})),
		},
	})
}
