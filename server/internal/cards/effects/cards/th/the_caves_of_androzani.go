package cardsth

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Caves of Androzani — Enchantment — Saga {3}{W} (Doctor Who):
//
//	"I — Put two stun counters on each of up to two target tapped
//	 creatures.
//	 II, III — For each non-Saga permanent, choose a counter on it. You may
//	 put an additional counter of that kind on that permanent.
//	 IV — Search your library for a Doctor card, reveal it, put it into your
//	 hand, then shuffle."
//
// II and III ask the Saga's controller about each non-Saga permanent that has
// a counter, in battlefield order, in one prompt that is the choice of the
// kind and the "you may" together (one option per kind, and a decline); a
// permanent with no counter has nothing to choose (CR 122.1). IV is a
// library search for a card with the Doctor subtype.
//
// No simplification.
func init() {
	gapB94AdditionalCounterAnswered = effects.OptionPickThen("option-pick/gap-caves-of-androzani-counter", func(ctx *effects.Context, r game.OptionPicked) error {
		if len(r.Carry) == 0 || r.Option == nil {
			return nil
		}
		id := r.Carry[0]
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			kinds := effects.GapB94CounterKinds(c)
			if r.Index >= 0 && r.Index < len(kinds) {
				if err := ctx.Game.AddCounterByForEffect(r.Chooser, id, kinds[r.Index], 1); err != nil {
					return err
				}
			}
		}
		return gapB94AskAdditionalCounter(ctx, r.Carry[1:])
	})
	const name = "The Caves of Androzani"
	stun := effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "put two stun counters on each of up to two target tapped creatures"),
		effects.TargetCreature("up to two target tapped creatures", effects.GapB94TappedCreature()).WithCount(0, 2),
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	effects.Register(effects.Spec{
		OracleID:     "5ce47fb9-4e4f-4419-af75-309d2743f5f9",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			stun,
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "for each non-Saga permanent, choose a counter on it; you may put an additional counter of that kind on it"), cavesOfAndrozaniCounters),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "for each non-Saga permanent, choose a counter on it; you may put an additional counter of that kind on it"), cavesOfAndrozaniCounters),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "search your library for a Doctor card"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.GapB94IsDoctorCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    name + " — a Doctor card",
						Source:    item.SourceCardID,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func cavesOfAndrozaniCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.HasSubtype("Saga") && len(effects.GapB94CounterKinds(c)) > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	return gapB94AskAdditionalCounter(ctx, ids)
}

// --- "choose a counter on it" prompts -------------------------------------

// gapB94AddAnotherCounterOptions are the options of "choose a counter on
// it. You may put an additional counter of that kind on it" for a
// permanent: one "put another <kind>" per kind it has, then a decline.
func gapB94AddAnotherCounterOptions(c game.Card) []game.ChoiceOption {
	var out []game.ChoiceOption
	for _, kind := range effects.GapB94CounterKinds(c) {
		out = append(out, game.ChoiceOption{Label: "Put another " + kind + " counter on " + c.Name})
	}
	return append(out, game.ChoiceOption{Label: "Don't put another counter on " + c.Name})
}

// gapB94AskAdditionalCounter asks the controller about the first permanent
// of `ids` that still has a counter (CR 608.2b: one that left or lost its
// counters is skipped), carrying the rest.
func gapB94AskAdditionalCounter(ctx *effects.Context, ids []uuid.UUID) error {
	for len(ids) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(ids[0])
		if !ok || len(effects.GapB94CounterKinds(c)) == 0 {
			ids = ids[1:]
			continue
		}
		return effects.PickOption{
			Player:   ctx.Controller(),
			Question: "The Caves of Androzani — choose a counter on " + c.Name + " and you may put another of that kind on it",
			Options:  gapB94AddAnotherCounterOptions(c),
			ThenKey:  gapB94AdditionalCounterAnswered,
			Carry:    slices.Clone(ids),
			WithItem: true,
		}.Apply(ctx)
	}
	return nil
}

// gapB94AdditionalCounterAnswered applies one answer and asks about the next
// permanent. An on-disk key: never renamed, never reused. Registered in
// the card's init, since the answer asks the next question, which names this key.
var gapB94AdditionalCounterAnswered game.OptionPickThen
