package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "19a2f0a0-9e68-4982-a5f5-b77d805befd7",
		Name:         "Demonic Pact",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			demonicPactTrigger(),
		},
	})
}

func demonicPactTrigger() game.TriggeredAbility {
	t := AtYourUpkeep(demonicPactLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeWithPurpose(ModeDoing("This enchantment deals 4 damage to any target and you gain 4 life.", TargetAny(),
			func(item *game.StackItem, ctx *Context, occ int) error {
				if t, ok := ModeTarget(ctx, occ); ok {
					if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
						return err
					}
				}
				return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
			}), ForTargets(DamageToTarget(0, 4))),
		ModeDoing("Target opponent discards two cards.", TargetPlayer("target opponent", Opponent()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
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
		ModeDoing("Draw two cards.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
			}),
		ModeDoing("You lose the game.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return LoseTheGame{Player: item.Controller}.Apply(ctx)
			}),
	)
	return t
}
