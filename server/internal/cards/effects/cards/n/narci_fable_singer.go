package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Narci, Fable Singer — Legendary Creature — Human Bard {1}{W}{B}{G}, 3/3
// (Commander Masters):
//
//	"Lifelink
//	 Whenever you sacrifice an enchantment, draw a card.
//	 Whenever the final chapter ability of a Saga you control resolves, each
//	 opponent loses X life and you gain X life, where X is that Saga's mana
//	 value."
//
// Lifelink is printed data. The sacrifice trigger reads the permanent as it
// was sacrificed (CR 603.10a), so an enchantment creature counts. The
// chapter trigger is Tom Bombadil's: it watches the resolve event of an
// ability that was put on the stack for the Saga's final chapter (CR 714.2d,
// 714.2e), for a Saga you control (the controller it had when the chapter
// triggered once it has left). X is that Saga's mana value, read as the
// trigger is put on the stack (the Saga may have sacrificed itself to its
// own chapter) and carried on the item; a Saga with {X} in its cost has
// X = 0 (CR 202.3e).
//
// No simplification.
func init() {
	chapter := effects.OnAny([]game.EventKind{game.EventResolve}, narciFinalChapterResolved,
		"Narci, Fable Singer — each opponent loses X life and you gain X life, X that Saga's mana value", narciDrain)
	chapter.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Narci, Fable Singer — each opponent loses X life and you gain X life, X that Saga's mana value")
		if tc, ok := g.ResolvingAbilityTriggerForEffect(ev.ResolvedStackItemID); ok {
			if saga, ok := g.LookupCardForEffect(tc.Event.CardID); ok {
				if mv, ok := saga.ParsedManaValue(); ok {
					item.Params.Amount = mv
				}
			}
		}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "61bbf710-4eb4-4e11-be6e-700a0c6a41ee",
		Name:            "Narci, Fable Singer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, narciSacrificedAnEnchantment,
				"Narci, Fable Singer — draw a card", effects.Do(effects.DrawCards{N: 1})),
			chapter,
		},
	})
}

// narciSacrificedAnEnchantment is "whenever you sacrifice an enchantment".
func narciSacrificedAnEnchantment(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.LeftAsType(ev, c, "enchantment")
}

// narciFinalChapterResolved: the ability now resolving is the final chapter
// ability of a Saga Narci's controller controls.
func narciFinalChapterResolved(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID != uuid.Nil {
		return false // a spell resolving
	}
	tc, ok := g.ResolvingAbilityTriggerForEffect(ev.ResolvedStackItemID)
	if !ok || tc.Event.Kind != game.EventSagaChapter {
		return false
	}
	saga, ok := g.LookupCardForEffect(tc.Event.CardID)
	if !ok || !game.IsSaga(saga) {
		return false
	}
	final := game.SagaFinalChapter(saga)
	if final <= 0 || tc.Event.Amount != final {
		return false
	}
	controller := tc.Event.Actor
	if z := g.FindCardZoneForEffect(saga.InstanceID); z != nil && z.Kind == game.ZoneBattlefield {
		controller = saga.Controller
	}
	return controller == source.Controller
}

// narciDrain is the chapter trigger's body: X is the amount stamped on the
// item as it was put on the stack.
func narciDrain(g *game.Game, item *game.StackItem) error {
	x := item.Params.Amount
	if x <= 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Amount: x}.Apply(ctx)
}
