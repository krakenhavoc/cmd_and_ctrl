package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wedding of River Song — Sorcery {2}{W} (Doctor Who):
//
//	"Draw two cards, then you may exile a nonland card from your hand with a
//	 number of time counters on it equal to its mana value. Then target opponent
//	 does the same. Cards exiled this way that don't have suspend gain suspend.
//	 Time travel. (For each suspended card you own and each permanent you control
//	 with a time counter on it, you may add or remove a time counter.)"
//
// In order: the draw; the caster's optional pick of a nonland card from their
// hand, exiled with the time counters and suspend (Amy's Home's pick and
// GainSuspendExileTail, CR 702.62a); then the target opponent's same pick (their
// own hand, their own card, counters equal to its mana value); then the caster
// time travels (CR 701.56a). Each step waits for the prompt the step before asks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ff91cc1-2233-4ceb-91ba-6a49c7147f32",
		Name:         "The Wedding of River Song",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return weddingAsk(ctx, ctx.Controller(), weddingControllerExile)
			}),
		),
	})
}

// weddingAsk asks `chooser` for a nonland card from their hand to exile (up to
// one); with none to offer there is nothing to ask and the next step runs.
func weddingAsk(ctx *effects.Context, chooser uuid.UUID, then game.ChooseCardsThen) error {
	g := ctx.Game
	p := g.PlayerByIDForEffect(chooser)
	var cards []uuid.UUID
	if p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if !c.IsLand() {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	if len(cards) == 0 {
		if chooser == ctx.Controller() {
			return weddingNext(ctx)
		}
		return weddingFinish(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  chooser,
		Source:   ctx.Source(),
		Question: "The Wedding of River Song — you may exile a nonland card from your hand with time counters equal to its mana value",
		Cards:    cards,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  then,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// weddingExile exiles the picked card with time counters equal to its mana
// value and suspend, as Amy's Home does, and returns whether it was asked.
func weddingExile(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ctx.Game.ExileCardKeyedForEffect(id, game.KeyedTail{
		Ref:    game.GainSuspendExileTail,
		Player: r.Chooser,
		N:      c.ManaValue(),
	})
}

// weddingNext asks the target opponent, once the caster has answered.
func weddingNext(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return weddingAsk(ctx, t.ID, weddingOpponentExile)
		}
	}
	return weddingFinish(ctx)
}

// weddingFinish is the last sentence: the caster time travels.
func weddingFinish(ctx *effects.Context) error { return effects.TimeTravel{}.Apply(ctx) }

// The two picks. On-disk keys: never renamed, never reused. Registered in init:
// the first answer asks the second question.
var weddingControllerExile, weddingOpponentExile game.ChooseCardsThen

func init() {
	weddingControllerExile = effects.AskedPick("choose-cards/the-wedding-of-river-song-you", func(ctx *effects.Context, r game.CardsChosen) error {
		if err := weddingExile(ctx, r); err != nil {
			return err
		}
		return weddingNext(ctx)
	})
	weddingOpponentExile = effects.AskedPick("choose-cards/the-wedding-of-river-song-opponent", func(ctx *effects.Context, r game.CardsChosen) error {
		if err := weddingExile(ctx, r); err != nil {
			return err
		}
		return weddingFinish(ctx)
	})
}
