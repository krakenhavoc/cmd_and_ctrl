package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emperor of Bones — Creature — Skeleton Noble {1}{B}, 2/2 (Modern Horizons 3):
//
//	"At the beginning of combat on your turn, exile up to one target card from
//	 a graveyard.
//	 {1}{B}: Adapt 2.
//	 Whenever one or more +1/+1 counters are put on this creature, put a
//	 creature card exiled with this creature onto the battlefield under your
//	 control with a finality counter on it. It gains haste. Sacrifice it at the
//	 beginning of the next end step."
//
// The exile is linked to this object (CR 607.2a); the counter trigger asks
// which creature card when more than one is linked.
func init() {
	exile := On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
		"Emperor of Bones — exile up to one target card from a graveyard", gapB49EmperorExile)
	exile.Targets = TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)
	Register(Spec{
		OracleID:     "cb70a360-898e-443a-b053-38e35ea2baa2",
		Name:         "Emperor of Bones",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{B}"),
			Effect:  adaptTwo,
		}},
		Triggered: []game.TriggeredAbility{
			exile,
			On(game.EventCounterPlaced, gapB49CountersPlacedOnThis,
				"Emperor of Bones — put a creature card exiled with it onto the battlefield with a finality counter; it gains haste; sacrifice it at the beginning of the next end step",
				gapB49EmperorReanimate),
		},
	})
}

func gapB49EmperorExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.ExileLinkedKeyedForEffect(link, []uuid.UUID{t.ID}, game.KeyedTail{})
	}
	return nil
}

func gapB49EmperorReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	link, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(link)) {
		if c.IsCreature() && !c.IsToken() {
			cands = append(cands, c.InstanceID)
		}
	}
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return gapB49EmperorPut(ctx, cands)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Emperor of Bones — put a creature card exiled with it onto the battlefield",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  gapB49EmperorPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// gapB49EmperorPick is the chosen card. An on-disk key: never renamed, never reused.
var gapB49EmperorPick = askedPick("choose-cards/emperor-of-bones-reanimate", func(ctx *Context, r game.CardsChosen) error {
	return gapB49EmperorPut(ctx, r.Picked)
})

func gapB49EmperorPut(ctx *Context, ids []uuid.UUID) error {
	return putCardsFromZoneTogetherWith(ctx, game.ZoneExile, ids, game.ZoneEntryOptions{
		Controller:         ctx.Controller(),
		EntersWithCounters: map[string]int{game.CounterFinality: 1},
	}, gapB49EmperorEntered, 0)
}

// gapB49EmperorEntered gives the returned card haste and schedules its
// sacrifice at the beginning of the next end step.
var gapB49EmperorEntered = CardTail("gap-b49/emperor-of-bones-entered", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Emperor of Bones — haste"}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: "Emperor of Bones — sacrifice it",
			Cards: []uuid.UUID{id},
			Body:  sacrificeListedCardsBody,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
