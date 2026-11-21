package effects

import (
	"github.com/google/uuid"

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
	const name = "The Caves of Androzani"
	stun := ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "put two stun counters on each of up to two target tapped creatures"),
		TargetCreature("up to two target tapped creatures", gapB94TappedCreature()).WithCount(0, 2),
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	Register(Spec{
		OracleID:     "5ce47fb9-4e4f-4419-af75-309d2743f5f9",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			stun,
			ChapterTrigger(2, SagaChapterLabel(name, 2, "for each non-Saga permanent, choose a counter on it; you may put an additional counter of that kind on it"), cavesOfAndrozaniCounters),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "for each non-Saga permanent, choose a counter on it; you may put an additional counter of that kind on it"), cavesOfAndrozaniCounters),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "search your library for a Doctor card"),
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: gapB94IsDoctorCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    name + " — a Doctor card",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

func cavesOfAndrozaniCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.HasSubtype("Saga") && len(gapB94CounterKinds(c)) > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	return gapB94AskAdditionalCounter(ctx, ids)
}
