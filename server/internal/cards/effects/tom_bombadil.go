package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tom Bombadil — Legendary Creature — God Bard {W}{U}{B}{R}{G}, 4/4
// (The Hobbit Commander):
//
//	"As long as there are four or more lore counters among Sagas you
//	 control, Tom Bombadil has hexproof and indestructible.
//	 Whenever the final chapter ability of a Saga you control resolves,
//	 reveal cards from the top of your library until you reveal a Saga
//	 card. Put that card onto the battlefield and the rest on the bottom
//	 of your library in a random order. This ability triggers only once
//	 each turn."
//
// The keywords are two self-only layer-6 grants gated on the board, read
// live: the lore counters on every Saga Tom's controller controls are
// added up, so two Sagas on chapter II are four.
//
// The trigger watches EventResolve, which an ability emits as it begins to
// resolve and a fizzled one never emits (CR 608.2b: an ability whose
// targets are all illegal does not resolve). The resolving item is read
// through game.ResolvingAbilityTriggerForEffect (#4505): a chapter ability
// was put on the stack for an EventSagaChapter carrying its chapter
// number, and it is the Saga's final chapter ability when that number is
// the Saga's final chapter number (CR 714.2d, 714.2e). "A Saga you
// control" is the Saga's controller while it is still on the battlefield,
// and the controller it had when the chapter triggered once it has left.
// Once each turn is the per-turn trigger record (b11TriggeredThisTurn),
// as Elrond, Moon-Reader reads it.
//
// The reveal is the reveal-until family's whole sentence
// (RevealUntilThenPutOntoBattlefield): no Saga card in the library
// reveals all of it and bottoms every card in a random order.
//
// No simplification. The Hobbit (#4454), seam #4505.
const (
	tomBombadilOracleID = "89a2d5bc-df06-47e8-8b4a-6d1f0a63c78d"
	tomBombadilLabel    = "Tom Bombadil — reveal until a Saga card and put it onto the battlefield"
)

func init() {
	Register(Spec{
		OracleID:     tomBombadilOracleID,
		Name:         "Tom Bombadil",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(tomFourLoreCounters, "hexproof"),
			KeywordGrant(tomFourLoreCounters, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventResolve}, tomFinalChapterResolved, tomBombadilLabel,
				Do(RevealUntilThenPutOntoBattlefield{
					Match:  game.IsSaga,
					Reason: "Tom Bombadil — reveal until a Saga card",
				})),
		},
	})
}

// tomFourLoreCounters is "as long as there are four or more lore counters
// among Sagas you control", for Tom himself.
func tomFourLoreCounters(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.InstanceID != source.InstanceID {
		return false
	}
	lore := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && game.IsSaga(c) {
			lore += c.Counters[game.CounterLore]
		}
	}
	return lore >= 4
}

// tomFinalChapterResolved: the ability now resolving is the final chapter
// ability of a Saga Tom's controller controls, and Tom has not triggered
// yet this turn.
func tomFinalChapterResolved(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
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
	return controller == source.Controller && !b11TriggeredThisTurn(g, source.InstanceID, tomBombadilLabel)
}
