package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lethal Vapors — Enchantment {2}{B}{B}:
//
//	"Whenever a creature enters, destroy it.
//	 {0}: Destroy this enchantment. You skip your next turn. Any player
//	 may activate this ability."
//
// The trigger is an ordinary one that goes on the stack and can be
// answered. It watches any creature entering, card or token, whoever
// controls it, and destroys exactly that object: one that left (or came
// back as a new object) before the trigger resolved is not destroyed
// (CR 400.7).
//
// The second line is an any-player row (CR 602.2, ADR 0106 §1). "You" is
// whoever activated it (CR 109.5), so an opponent who destroys the
// enchantment skips THEIR next turn, not its controller's. The skip is a
// separate sentence from the destruction (CR 614.10, ADR 0188 Decision 1):
// it happens even when the enchantment was already gone by the time the
// ability resolved, and it is a debt spent when the activator's next turn
// would begin, so it never cuts short a turn in progress.
//
// No purpose for the bot: a row with none is never chosen by a bot
// reaching across the table, and the controller has no reason to pay a
// turn to remove its own enchantment. The row still declares what it
// answers, which smart auto-pass reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fe9650d-6c9b-4988-9ea4-79ba0d55a77b",
		Name:         "Lethal Vapors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33AnyCreatureEntered(ev, g)
			}, "Lethal Vapors — destroy it", lethalVaporsDestroyEntered),
		},
		Activated: []ActivatedAbility{{
			Label: "{0}: Destroy this enchantment. You skip your next turn. Any player may activate this ability.",
			// No SkipsTurns here on purpose: this is an any-player row, the
			// skip falls on the activator, and Register refuses the field
			// there (ADR 0188 Decision 10, ADR 0106 decision 2).
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{0}"),
			AnyPlayer: true,
			Effect:    lethalVaporsDestroyAndSkip,
		}},
	})
}

// lethalVaporsDestroyEntered destroys the creature whose entry triggered
// the ability, if it is still that object on the battlefield.
func lethalVaporsDestroyEntered(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	entered := item.Trigger.Event.CardID
	if !onBattlefield(g, entered) {
		return nil
	}
	return DestroyTarget{Target: entered}.Apply(NewContext(g, item))
}

// lethalVaporsDestroyAndSkip is the second line's effect: destroy this
// enchantment, then the activator skips their next turn.
func lethalVaporsDestroyAndSkip(g *game.Game, item *game.StackItem) error {
	if err := destroyThisPermanent(false)(g, item); err != nil {
		return err
	}
	return SkipYourNextTurns{}.Apply(NewContext(g, item))
}
