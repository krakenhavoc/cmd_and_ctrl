package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// demonicPactLabel is the trigger's stack label, and with it the key
// its "hasn't been chosen" memory is kept under.
const demonicPactLabel = "Demonic Pact — choose one that hasn't been chosen"

// Demonic Pact — Enchantment {2}{B}{B}:
//
//	"At the beginning of your upkeep, choose one that hasn't been
//	 chosen —
//	 • This enchantment deals 4 damage to any target and you gain 4
//	   life.
//	 • Target opponent discards two cards.
//	 • Draw two cards.
//	 • You lose the game."
//
// The card ADR 0097 was written around. ChooseOneNotChosen keeps the
// memory on this OBJECT for as long as it is on the battlefield, so
// the fourth upkeep offers "You lose the game" alone and a mandatory
// trigger must take it (the 2015-06-22 ruling: "if the fourth mode is
// the only one remaining, you must choose it"). The rulings' other
// three points are the engine's too:
//
//   - a mode chosen for an instance that is countered "still counts as
//     being chosen" — the choice is recorded when it is made;
//   - "it doesn't matter who has chosen any particular mode" — the
//     memory follows the enchantment, not its controller, so a Pact
//     given to an opponent (Harmless Offering) offers them only the
//     modes left, and "you lose the game" is then theirs;
//   - a second Demonic Pact, or this one after it leaves and returns,
//     is a new object that may choose any mode.
//
// "You gain 4 life" happens even if the damage is prevented, because
// the sentence is not "if you do"; a target that became illegal takes
// the whole instance with it (CR 608.2b). "Target opponent discards
// two cards" is the opponent's choice of cards (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19a2f0a0-9e68-4982-a5f5-b77d805befd7",
		Name:         "Demonic Pact",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			demonicPactTrigger(),
		},
	})
}

func demonicPactTrigger() game.TriggeredAbility {
	t := effects.AtYourUpkeep(demonicPactLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosen(
		// Two steps, so the life gain waits for a prompt the damage asks
		// (Khârn the Betrayer's question, a CR 616 order), CR 608.2c
		// (#4019).
		effects.ModeWithPurpose(effects.ModeSteps("This enchantment deals 4 damage to any target and you gain 4 life.", effects.TargetAny(),
			effects.Instruction(demonicPactDamage), effects.GainLife{Amount: 4}), effects.ForTargets(effects.DamageToTarget(0, 4))),
		effects.ModeDoing("Target opponent discards two cards.", effects.TargetPlayer("target opponent", effects.Opponent()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   t.ID,
					Source:   item.SourceCardID,
					N:        2,
					Question: "Demonic Pact — discard two cards",
				})
				return nil
			}),
		effects.ModeDoing("Draw two cards.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
			}),
		effects.ModeDoing("You lose the game.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.LoseTheGame{Player: item.Controller}.Apply(ctx)
			}),
	)
	return t
}

// demonicPactDamage is the first bullet's damage: 4 to its target, when
// that target is still legal (CR 608.2b).
func demonicPactDamage(ctx *effects.Context) error {
	t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
}
