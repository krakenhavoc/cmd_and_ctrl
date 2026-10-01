package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armageddon Clock — Artifact {6}:
//
//	"At the beginning of your upkeep, put a doom counter on this
//	 artifact.
//	 At the beginning of your draw step, this artifact deals damage
//	 equal to the number of doom counters on it to each player.
//	 {4}: Remove a doom counter from this artifact. Any player may
//	 activate this ability but only during any upkeep step."
//
// ADR 0106 PR 6 (#1793). Each line is an existing shape:
//
//   - The upkeep trigger puts the counter on (AtYourUpkeep). "Your" is
//     the Clock's controller.
//   - The draw-step trigger watches EventBeginDrawStep for its
//     controller's draw step, which fires after the turn-based draw
//     (CR 504.1); the trigger goes on the stack when the active player
//     next gets priority (CR 504.2). The amount is read as the ability
//     resolves (CR 608.2h), off the Clock as it last existed if it has
//     left, and the Clock deals that much to every player still in the
//     game, its controller included.
//   - The any-player row (CR 602.2, 602.1b). "Only during any upkeep
//     step" is an instruction about activating it (CR 602.1b), read
//     with DuringStep(StepUpkeep): any player's upkeep, not only the
//     activator's. The {4} is the activator's to pay (CR 602.1a).
//     Removing the counter is the EFFECT, so a Clock with no doom
//     counters can still be activated and then does nothing (CR 609.3).
//
// No purpose for the bot: removing a counter protects every player,
// and no ActivationPurpose field says so, so a bot never reaches for
// another player's Clock.
//
// No simplification.
func init() {
	const doom = "doom"
	Register(Spec{
		OracleID:     "70d90ef4-0cda-405f-abf1-734fa909efa6",
		Name:         "Armageddon Clock",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Armageddon Clock — put a doom counter", putACounterOnThis(doom)),
			On(game.EventBeginDrawStep, ByYou, "Armageddon Clock — damage to each player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					return b38DamageEachPlayer(ctx, info.Counters[doom])
				}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{4}: Remove a doom counter from this artifact. Any player may activate this ability but only during any upkeep step.",
			Cost:      game.AbilityCost{Mana: "{4}"},
			AnyPlayer: true,
			Condition: DuringStep(game.StepUpkeep),
			Effect:    removeACounterFromThis(doom),
		}},
	})
}
