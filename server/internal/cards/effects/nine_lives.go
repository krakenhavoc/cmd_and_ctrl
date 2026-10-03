package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nine Lives — Enchantment {1}{W}{W}:
//
//	"Hexproof
//	 If a source would deal damage to you, prevent that damage and put an incarnation counter on this enchantment.
//	 When there are nine or more incarnation counters on this enchantment, exile it.
//	 When this enchantment leaves the battlefield, you lose the game."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect runs
// once per SOURCE in a damage instance (PreventDamageASourceWouldDeal):
// "if more than one source deals damage to you at once, prevent the damage
// from each of them and put that many incarnation counters", even past
// nine, and under damage that can't be prevented the counter still goes on
// (CR 615.12; the rulings). The nine-counter clause is a CR 603.8 state
// trigger (ADR 0107 PR 1), so it triggers once however far past nine the
// counters go.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "236e1f57-7ef5-455a-82a6-8ff6b85d8849",
		Name:            "Nine Lives",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Replacements: []game.ReplacementEffect{
			PreventDamageASourceWouldDeal(PreventionStatic{
				To:    ToYou,
				Then:  nineLivesCounterBody,
				Label: "Nine Lives — prevent the damage and put an incarnation counter on it",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisHasAtLeast(nineLivesCounter, 9, "Nine Lives — exile it", ExileThisThen(nil)),
			WhenThisLeaves("Nine Lives — you lose the game", Do(LoseTheGame{})),
		},
	})
}

const nineLivesCounter = "incarnation"

// The additional effect: one incarnation counter per application.
var nineLivesCounterBody = game.DelayedBody("nine-lives/incarnation-counter", nineLivesAddCounter)

func nineLivesAddCounter(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	if _, ok := followUpThis(g, item); !ok {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, item.SourceCardID, nineLivesCounter, 1)
}
