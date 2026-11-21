package cardsp

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parnesse, the Subtle Brush — Legendary Creature — Vampire Wizard
// {2}{U}{B}{R}, 4/4 (Streets of New Capenna Commander, #5491):
//
//	"Whenever you or a permanent you control becomes the target of a spell
//	 or ability an opponent controls, counter that spell or ability unless
//	 that player pays 4 life.
//	 Whenever you copy a spell, up to one target opponent may also copy
//	 that spell. They may choose new targets for that copy."
//
// The first trigger is ward's shape for a player as well as a permanent
// (Surrak's EventBecomesTarget watch, fired once per target slot as the
// targets are chosen, CR 115.3, so it resolves before the spell or ability
// that did the targeting): an opponent targeting you or one of your
// permanents is asked to pay 4 life or have it countered. Paying is a cost
// (CR 118.3) and needs a life total of at least 4 (CR 119.4); one who
// cannot pay is not asked. The targeting spell or ability is named by the
// event, so it is countered by its stack ID, and one already gone is
// ignored.
//
// The second watches the engine's spell-copied event (CR 707.10), whose
// actor is the copy's controller, so a copy an opponent makes does not
// trigger it again, and a card cast from a copy's zone (CR 707.12) is cast,
// not copied. The target is optional ("up to one"); the chosen opponent is
// asked "you may" as the ability resolves, and a yes makes a copy that
// opponent controls (CR 707.10c lets them choose new targets). The spell is
// named, not targeted, so one that has left the stack is copied from
// last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6040e8d-dac3-4bed-8d1d-d7d44aa10c86",
		Name:         "Parnesse, the Subtle Brush",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, s82SncParnesseYouAreTargeted,
				"Parnesse, the Subtle Brush — counter that spell or ability unless that player pays 4 life",
				s82SncParnesseCounterUnlessLife),
			effects.Targeting(effects.On(game.EventSpellCopied, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventSpellCopied && ev.Actor == source.Controller
			}, "Parnesse, the Subtle Brush — up to one target opponent may also copy that spell",
				s82SncParnesseOpponentCopies), effects.TargetPlayer("up to one target opponent", effects.Opponent()).WithCount(0, 1)),
		},
	})
}

// s82SncParnesseYouAreTargeted is "you or a permanent you control becomes
// the target of a spell or ability an opponent controls".
func s82SncParnesseYouAreTargeted(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.Actor == uuid.Nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	if ev.CardID == uuid.Nil {
		return ev.Target == you
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.OnBattlefield(g, ev.CardID) && c.Controller == you
}

// s82SncParnesseCounterUnlessLife is the first trigger's resolution: the
// targeting spell or ability's controller pays 4 life or it is countered.
func s82SncParnesseCounterUnlessLife(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	targeting, payer := ev.StackItemID, ev.Actor
	if g.StackItemForEffect(targeting) == nil {
		return nil
	}
	if p := g.PlayerByIDForEffect(payer); p == nil || p.Life < 4 {
		return effects.WardCounter(g, targeting)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      payer,
		Source:       item.SourceCardID,
		Question:     fmt.Sprintf("Parnesse, the Subtle Brush — pay %d life, or it is countered", 4),
		AcceptLabel:  "Pay 4 life",
		DeclineLabel: "Let it be countered",
		LifeCost:     4,
	}, effects.WardPayLifeAnswer, []uuid.UUID{targeting})
	return err
}

// s82SncParnesseOpponentCopies asks the chosen opponent whether they copy
// the spell that was copied.
func s82SncParnesseOpponentCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return effects.MayChoice{
			Player:   t.ID,
			Question: "Parnesse, the Subtle Brush — copy that spell?",
			Then:     s82SncParnesseCopyAnswer,
			Carry:    []uuid.UUID{item.Trigger.Event.Source},
		}.Apply(ctx)
	}
	return nil
}

// s82SncParnesseCopyAnswer makes the opponent's copy. An on-disk key: never
// renamed, never reused.
var s82SncParnesseCopyAnswer = effects.MayYes("confirm/parnesse-the-subtle-brush-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	err := ctx.Game.CopyLastKnownSpellForEffect(r.Carry[0], r.Chooser, true, nil)
	if err == game.ErrCardNotFound {
		return nil
	}
	return err
})
