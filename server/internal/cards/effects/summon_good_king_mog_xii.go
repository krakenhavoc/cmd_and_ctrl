package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Good King Mog XII — Enchantment Creature — Saga Moogle {4}{W},
// 4/4:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Create two 1/2 white Moogle creature tokens with lifelink.
//	 II, III — Whenever you cast a noncreature spell this turn, create a
//	 token that's a copy of a non-Saga token you control.
//	 IV — Put two +1/+1 counters on each other Moogle you control.
//	 Flying, lifelink"
//
// Chapters II and III schedule a delayed trigger that watches every
// noncreature spell the Saga's controller casts for the rest of the turn
// (CR 603.7b, DelayedOnEvent with Repeats), so a chapter that triggers
// on a turn in which nothing more is cast does nothing. The copy is
// populate's choice widened to any non-Saga token (a Treasure or a Clue
// counts); with every candidate alike it is not a choice, otherwise the
// controller picks. Chapter IV reads each other Moogle as it resolves,
// the Saga's tokens and any other Moogle alike.
//
// No simplification.
func init() {
	const name = "Summon: Good King Mog XII"
	copyBoth := func(g *game.Game, item *game.StackItem) error {
		return DelayedOnEvent{
			Label:     name + " — create a token that's a copy of a non-Saga token you control",
			On:        []game.EventKind{game.EventCast},
			Condition: mogNoncreatureCastCondition,
			Body:      mogCopyTokenBody,
			Repeats:   true,
		}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:        "2148820c-85b6-4598-adf2-10873001a779",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create two 1/2 white Moogle creature tokens with lifelink"),
				Do(CreateToken{Template: TokenCard("1/2 white Moogle with lifelink"), N: 2})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "whenever you cast a noncreature spell this turn, create a token that's a copy of a non-Saga token you control"), copyBoth),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "whenever you cast a noncreature spell this turn, create a token that's a copy of a non-Saga token you control"), copyBoth),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "put two +1/+1 counters on each other Moogle you control"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item).asGroupMember()
					// "Moogle" is a word of the source's text (CR 612.2, #3984).
					moogle := ctx.TextSubtype("Moogle")
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || c.InstanceID == item.SourceCardID || !c.HasSubtype(moogle) {
							continue
						}
						if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

var (
	// "Whenever you cast a noncreature spell": an EventCast by the
	// delayed trigger's controller of a spell that is not a creature.
	mogNoncreatureCastCondition = game.DelayedCondition("good-king-mog/you-cast-a-noncreature-spell", mogNoncreatureCast)

	// "Create a token that's a copy of a non-Saga token you control."
	mogCopyTokenBody = game.SimpleDelayedBody("good-king-mog/copy-a-non-saga-token", mogCopyToken)
)

func mogNoncreatureCast(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !spell.IsCreature()
}

func mogCopyToken(g *game.Game, item *game.StackItem) error {
	return Populate{
		Match: And(IsTokenPredicate(), Not(HasSubtype("Saga"))),
	}.Apply(NewContext(g, item))
}
