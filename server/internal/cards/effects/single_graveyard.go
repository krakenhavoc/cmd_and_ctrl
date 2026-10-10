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
		// ADR 0142 sweep ruling 9: graveyard hate is restrict.
		Purpose: game.Purpose{Answers: game.AnswerRestrict},
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
	return exileTargetCardsThen(ctx, nil)
}

// ExileTargetCards exiles every card target that is still legal as the
// spell or ability resolves (CR 608.2b), in one simultaneous move. A
// target that left its graveyard in response is skipped, and the rest
// still go. Decompose, Carrion Beetles, Digsite Conservator.
func ExileTargetCards(g *game.Game, item *game.StackItem) error {
	return exileTargetCardsThen(NewContext(g, item), nil)
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
func exileTargetCardsThen(ctx *Context, then func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error) error {
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
func exileCardsThen(ctx *Context, cards []uuid.UUID, then func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error) error {
	var ids []uuid.UUID
	wasCreature := map[uuid.UUID]bool{}
	for _, id := range cards {
		card, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		ids = append(ids, id)
		wasCreature[id] = card.IsCreature()
	}
	item := ctx.Item
	return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		if then == nil {
			return nil
		}
		return then(NewContext(g, item), exiled, wasCreature)
	})
}
