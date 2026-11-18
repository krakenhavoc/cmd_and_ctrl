package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vanish into Memory's two halves. An on-disk key each: never renamed,
// never reused.
var (
	// vanishIntoMemoryExiled is the exile settled: draw cards equal to
	// the creature's power, then schedule its return.
	vanishIntoMemoryExiled = CardTail("then/vanish-into-memory-exiled", vanishIntoMemoryAfterExile)

	// vanishIntoMemoryReturnBody is the delayed trigger: return the card,
	// and if it returned, discard cards equal to its toughness.
	vanishIntoMemoryReturnBody = game.DelayedBody("return/vanish-into-memory", vanishIntoMemoryReturn)

	// vanishIntoMemoryEntered is the card back: the discard.
	vanishIntoMemoryEntered = TogetherTail("vanish-into-memory/entered", func(ctx *Context, entered []uuid.UUID, n int) error {
		if len(entered) == 0 || n <= 0 {
			return nil
		}
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        n,
			Question: "Vanish into Memory — discard cards equal to its toughness",
		})
		return nil
	})
)

// vanishIntoMemoryResolve is the spell: the creature's power and
// toughness are read now, as it last existed (CR 608.2h), and ride the
// exile's tail.
func vanishIntoMemoryResolve(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	tail := ctx.ItemTail(vanishIntoMemoryExiled)
	tail.Nums = []int{c.CurrentPower(), c.CurrentToughness()}
	return ExileTarget{Target: id, ThenTail: tail}.Apply(ctx)
}

func vanishIntoMemoryAfterExile(ctx *Context, o game.TailOutcome) error {
	power, toughness := 0, 0
	if len(o.Tail.Nums) == 2 {
		power, toughness = o.Tail.Nums[0], o.Tail.Nums[1]
	}
	if power > 0 {
		if err := (DrawCards{N: power}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Vanish into Memory — return that card to the battlefield under its owner's control, then discard",
		Cards:              []uuid.UUID{o.IDs[0]},
		Body:               vanishIntoMemoryReturnBody,
		Params:             game.EffectParams{Object: game.ObjectRef{ID: o.IDs[0], Epoch: c.ObjectEpoch}, Amount: toughness},
	}.Apply(ctx)
}

func vanishIntoMemoryReturn(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if _, ok := objectStillInAPublicZone(g, p.Object); !ok {
		return nil
	}
	return ReturnFromExileTogether{
		Targets:  []uuid.UUID{p.Object.ID},
		ThenTail: vanishIntoMemoryEntered,
		ThenN:    p.Amount,
	}.Apply(NewContext(g, item))
}
