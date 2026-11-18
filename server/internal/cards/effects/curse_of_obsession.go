package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of Obsession — Enchantment — Aura Curse {4}{R}:
//
//	"Enchant player
//	 At the beginning of enchanted player's draw step, that player draws
//	 two additional cards.
//	 At the beginning of enchanted player's end step, that player
//	 discards their hand."
//
// Two enchant-player triggers, each reading the player whose step it is
// off the event and firing only when that is the player the Curse is
// attached to. Howling Mine's shape for the draw (the turn-based draw
// has already happened, so these are the "additional" cards); the
// discard is the whole hand, the player's own cards, with no choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca4b4bf2-ed34-444c-9088-dd1ea488ec4e",
		Name:         "Curse of Obsession",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginDrawStep, curseOfObsessionTheirStep,
				"Curse of Obsession — they draw two additional cards",
				Do(Instruction(func(ctx *Context) error {
					return ctx.Game.DrawNForEffect(ctx.Item.Trigger.Event.Actor, 2)
				}))),
			On(game.EventBeginEndStep, curseOfObsessionTheirStep,
				"Curse of Obsession — they discard their hand",
				Do(Instruction(func(ctx *Context) error {
					_, err := discardWholeHand(ctx.Game, ctx.Item.Trigger.Event.Actor)
					return err
				}))),
		},
	})
}

// curseOfObsessionTheirStep is "at the beginning of enchanted player's
// step": the step is the attached player's own.
func curseOfObsessionTheirStep(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
}
