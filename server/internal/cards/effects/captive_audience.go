package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// captiveAudienceLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen" memory is kept under.
const captiveAudienceLabel = "Captive Audience — choose one that hasn't been chosen"

// Captive Audience — Enchantment {5}{B}{R}:
//
//	"This enchantment enters under the control of an opponent of your
//	 choice.
//	 At the beginning of your upkeep, choose one that hasn't been
//	 chosen —
//	 • Your life total becomes 4.
//	 • Discard your hand.
//	 • Each opponent creates five 2/2 black Zombie creature tokens."
//
// The card #1759 was filed for. The first line is ADR 0102's
// EntersUnderTheControlOfAnOpponentOfYourChoice: asked as the
// enchantment would enter, so it lands under the chosen opponent, who
// owns nothing about it but the control. The 2019-01-25 rulings are
// all the engine's:
//
//   - "Once Captive Audience is controlled by your opponent, its
//     ability triggers during that player's upkeep and that player
//     makes all choices for it" — "your" is the controller, and the
//     trigger is theirs (CR 603.3a).
//   - "That player is affected by its first two modes, and that
//     player's opponents create the tokens for its last mode" — the
//     caster included.
//   - "If all three modes have been chosen, Captive Audience's
//     triggered ability is removed from the stack with no effect" —
//     ADR 0097's ChooseOneNotChosen.
//   - The owner leaving takes it out of the game; the controller
//     leaving exiles it (CR 800.4a).
//
// "Your life total becomes 4" gains or loses the difference (CR
// 119.5), through the same helper Sorin Markov uses. "Discard your
// hand" takes every card, so there is no choice to make.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc25ca51-f351-4877-a092-525fb48524a1",
		Name:         "Captive Audience",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Captive Audience", game.ControlForHarm),
		},
		Triggered: []game.TriggeredAbility{
			captiveAudienceTrigger(),
		},
	})
}

func captiveAudienceTrigger() game.TriggeredAbility {
	t := AtYourUpkeep(captiveAudienceLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Your life total becomes 4.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return b31LifeBecomes(ctx, item.Controller, 4)
			}),
		ModeDoing("Discard your hand.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				_, err := discardWholeHand(ctx.Game, item.Controller)
				return err
			}),
		ModeDoing("Each opponent creates five 2/2 black Zombie creature tokens.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				for _, opp := range ctx.Opponents() {
					if err := (CreateToken{Controller: opp, Template: TokenCard("2/2 black Zombie"), N: 5}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
	)
	return t
}
