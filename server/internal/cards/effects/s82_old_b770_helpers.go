package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b770_helpers.go — shared shapes of S82 OLD batch 770 (Tempest: shadow,
// artifacts, graveyard recursion, mana and ramp, counterspells and bounce;
// #5628): the shadow creatures' triggers.

// S82OldB770ACreatureWithShadowDied is "whenever a creature with shadow dies"
// (Dauthi Ghoul): a creature put into a graveyard from the battlefield that had
// shadow as it last existed there (CR 603.10a), whoever controlled it.
func S82OldB770ACreatureWithShadowDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if _, ok := DiedCreature(ev, g); !ok {
		return false
	}
	had, known := ev.WasKeyword("shadow")
	return known && had
}

// s82OldB770MindripperAnswer is the "you may sacrifice it" of Dauthi Mindripper:
// on a yes the Mindripper is sacrificed, and the discard waits for the
// sacrifice to land. The defending player rides the Carry.
var s82OldB770MindripperAnswer = MayYes("confirm/s82-old-b770-dauthi-mindripper", func(ctx *Context, r game.ConfirmAnswer) error {
	defender := CarryAt(r, 0)
	if defender == uuid.Nil || !OnBattlefield(ctx.Game, ctx.Source()) || SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	tail := ctx.Tail(s82OldB770MindripperDiscard)
	tail.Carry = []uuid.UUID{defender}
	return SacrificePermanent{Target: ctx.Source(), ThenTail: tail}.Apply(ctx)
})

// s82OldB770MindripperDiscard is "If you do, defending player discards three
// cards": told whether the Mindripper really left the battlefield, and run only
// then. The defending player chooses what to discard (CR 701.9a).
var s82OldB770MindripperDiscard = CardTail("s82-old-b770/dauthi-mindripper-discard", func(ctx *Context, o game.TailOutcome) error {
	if !OneWentThisWay(o) || len(o.Tail.Carry) == 0 || !DefendingPlayerStillIn(ctx.Game, o.Tail.Carry[0]) {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: o.Tail.Carry[0], Source: ctx.Source(), N: 3})
	return nil
})

// S82OldB770MindripperTrigger is Dauthi Mindripper's second ability.
func S82OldB770MindripperTrigger() game.TriggeredAbility {
	const label = "Dauthi Mindripper — you may sacrifice it. If you do, defending player discards three cards"
	return WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
		return MayChoice{
			Question: label + "?",
			Then:     s82OldB770MindripperAnswer,
			Carry:    []uuid.UUID{defender},
		}.Apply(NewContext(g, item))
	})
}
