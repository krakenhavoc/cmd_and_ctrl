package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrodominance — Legendary Enchantment {B}{B}{B}:
//
//	"Skip your draw step.
//	 At the beginning of your end step, you may pay any amount of life.
//	 If you do, draw that many cards.
//	 Your maximum hand size is five.
//	 If a card or token would be put into your graveyard from anywhere,
//	 exile it instead."
//
// Four clauses, each on a seam that already exists, and one that needed
// #1941:
//
//   - The skip is Necropotence's replacement (SkipYourDrawStep).
//   - The end-step payment is the life form of the number prompt (ADR
//     0129's amendment of 2026-10-09): any amount from 0 to your life
//     total (CR 119.4), paid as the trigger resolves (CR 118.12), and
//     that many cards drawn. Paying nothing draws nothing.
//   - The hand size is Spec.HandSize (ADR 0113 §3).
//   - The exile is Forbidden Crypt's replacement (GraveyardBecomesExile
//     with YoursOnly): the graveyard a card or token would go to is its
//     owner's, so "your graveyard" is the cards and tokens you own.
//
// A bot pays for the cards that fill its hand to five, no more than its
// library holds and never below 10 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a10b3e35-8cc4-450e-9e30-0fce8df0fea4",
		Name:         "Necrodominance",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SkipYourDrawStep(),
			GraveyardBecomesExile{
				YoursOnly: true,
				Label:     "Necrodominance: exile instead of your graveyard",
			}.Build(),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Necrodominance — pay any amount of life, draw that many cards", necrodominanceEndStep),
		},
		HandSize: []game.HandSizeStatic{YourMaxHandSizeIs(5)},
	})
}

// necrodominanceEndStep is "you may pay any amount of life. If you do,
// draw that many cards."
func necrodominanceEndStep(g *game.Game, item *game.StackItem) error {
	return PayLifeAmount{
		Question: "Necrodominance — pay any amount of life to draw that many cards",
		Unit:     game.PayAmountCards,
		Goal:     necrodominanceGoal,
		Then: func(ctx *Context, paid int) error {
			if paid <= 0 {
				return nil
			}
			return DrawCards{Player: ctx.Controller(), N: paid}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}

// necrodominanceGoal is what a bot pays: the cards that fill its hand to
// the maximum of five, capped by its library and by 10 life kept.
func necrodominanceGoal(ctx *Context) int {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil {
		return 0
	}
	want := 5 - len(p.Hand.Cards)
	if lib := len(p.Library.Cards); want > lib {
		want = lib
	}
	return max(0, min(want, p.Life-10))
}
