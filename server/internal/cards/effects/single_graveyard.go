package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// single_graveyard.go — #1807 (ADR 0106 §5): the shared body behind
// "Exile up to <n> target cards from a single graveyard", the
// graveyard-hate sentence about two dozen cards print. The clause is
// UpToCardsFromASingleGraveyard (target_set.go); this is what it does.
//
// Append-only, like every mechanic file in this package.

// creatureCardsAmong counts the cards of `exiled` that were creature
// cards before they moved — "for each creature card exiled this way"
// (Waste Management). exileTargetCardsThen supplies both arguments.
func creatureCardsAmong(exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) int {
	n := 0
	for _, id := range exiled {
		if wasCreature[id] {
			n++
		}
	}
	return n
}

// anyWasCreature is "if at least one creature card was exiled this
// way" (Soul-Shackled Zombie, Kaya, Orzhov Usurper).
func anyWasCreature(exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) bool {
	return creatureCardsAmong(exiled, wasCreature) > 0
}

// upToNCardsFromASingleGraveyard is the clause with its printed wording
// spelled from the count: "up to three target cards from a single
// graveyard".
func upToNCardsFromASingleGraveyard(n int) *game.TargetSpec {
	return UpToCardsFromASingleGraveyard("up to "+numberWord(n)+" target cards from a single graveyard", n)
}

// ExileFromASingleGraveyardAbility is the activated ability "<cost>:
// Exile up to <n> target cards from a single graveyard." `costText` is
// the cost as printed ("{2}{B}, {T}"), `cost` the same cost as data.
// Carrion Beetles and Rag Dealer print it word for word; Famished
// Ghoul and Unlicensed Hearse with other costs and counts.
func ExileFromASingleGraveyardAbility(costText string, cost game.AbilityCost, n int) ActivatedAbility {
	return ActivatedAbility{
		Label: costText + ": Exile up to " + numberWord(n) + " target cards from a single graveyard.",
		// ADR 0142 sweep ruling 9, split from restrict by ADR 0009's
		// amendment of 2026-10-30 (#4734, Q5): graveyard hate is graveyard.
		Purpose: game.Purpose{Answers: game.AnswerGraveyard},
		Cost:    cost,
		Targets: upToNCardsFromASingleGraveyard(n),
		Effect:  ExileTargetCards,
	}
}

// WhenThisEntersExileFromASingleGraveyard is "When this creature
// enters, exile up to <n> target cards from a single graveyard" —
// Griffnaut Tracker, Arashin Sunshield. `effect` is ExileTargetCards
// for the bare sentence, or a body that reads what was exiled.
//
// With every graveyard empty the trigger still triggers and resolves
// doing nothing: an "up to" clause is never unfillable (CR 603.3d).
func WhenThisEntersExileFromASingleGraveyard(name string, n int, effect Effect) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(name+" — exile up to "+numberWord(n)+" target cards from a single graveyard", effect),
		upToNCardsFromASingleGraveyard(n))
}

// exileTargetCardsOnResolve is ExileTargetCards as a spell's OnResolve
// — Decompose, Scarab Feast, Rapid Decay, Shred Memory.
func exileTargetCardsOnResolve(_ *game.StackItem, ctx *Context) error {
	return exileTargetCardsThen(ctx, game.TailRef{})
}

// ExileTargetCards exiles every card target that is still legal as the
// spell or ability resolves (CR 608.2b), in one simultaneous move. A
// target that left its graveyard in response is skipped, and the rest
// still go. Decompose, Carrion Beetles, Digsite Conservator.
func ExileTargetCards(g *game.Game, item *game.StackItem) error {
	return exileTargetCardsThen(NewContext(g, item), game.TailRef{})
}

// exileTargetCardsThen is ExileTargetCards with a clause hanging off
// the cards that ACTUALLY reached exile — "if at least one creature
// card was exiled this way" (Soul-Shackled Zombie). `wasCreature` is
// read before the move, because the clause asks what each card was in
// the graveyard, and `exiled` comes from the batch's continuation,
// so a commander whose owner sent it to the command zone (CR 903.9)
// is not counted (#870, The Binding of the Titans).
//
// The context is rebuilt inside the continuation from the live *Game,
// the contract massEffect.apply explains.
func exileTargetCardsThen(ctx *Context, then game.TailRef) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		ids = append(ids, t.ID)
	}
	return exileCardsThen(ctx, ids, then)
}

// exileCardsThen is exileTargetCardsThen over a list of cards that are
// not targets — Waste Management's kicked "exile target player's
// graveyard", which still counts the creature cards exiled this way.
//
// #4253 slice 9: `then` is a registered tail (GraveyardExileTail), or the
// zero ref for none, so a table waiting on a CR 616 prompt in the middle
// of the exile is a restore point.
func exileCardsThen(ctx *Context, cards []uuid.UUID, then game.TailRef) error {
	var ids []uuid.UUID
	for _, id := range cards {
		if _, ok := ctx.Game.LookupCardForEffect(id); !ok {
			continue
		}
		ids = append(ids, id)
	}
	var t game.KeyedTail
	if then.Key() != "" {
		t = ctx.ItemTail(then)
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, t)
}

// GraveyardExileTail registers the clause hanging off a graveyard exile
// (exileTargetCardsThen, exileCardsThen), told the cards that reached exile
// and which of them are creature cards. A graveyard card keeps what it is
// on its way to exile, so the creature cards are read off the exiled
// cards.
func GraveyardExileTail(key string, body func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		wasCreature := make(map[uuid.UUID]bool, len(o.IDs))
		for _, id := range o.IDs {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				wasCreature[id] = c.IsCreature()
			}
		}
		return body(ctx, o.IDs, wasCreature)
	})
}
