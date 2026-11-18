package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boreas Charger — Creature — Pegasus {2}{W}, 2/1:
//
//	"Flying
//	 When this creature leaves the battlefield, choose an opponent who
//	 controls more lands than you. Search your library for a number of
//	 Plains cards equal to the difference, reveal those cards, put one of
//	 them onto the battlefield tapped and the rest into your hand, then
//	 shuffle."
//
// The opponent is chosen from those who control more lands than you, as
// the trigger resolves; with none, nothing happens (the choice, and so the
// search, is impossible). The difference is counted when the choice is
// answered. The search is for "Plains cards", any card with the Plains
// subtype, nonbasic ones included, up to that many (a stated quality may
// find fewer, CR 701.23b).
//
// The found cards are revealed and the library is shuffled, then they are
// taken from the top of it: the shuffle is not observable, and this way
// the cards go straight to their zones, one to the battlefield tapped, as
// the player chooses (PutFromLibraryOntoBattlefield, Max 1), and "the
// rest" to hand, never passing through the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f6dae324-cc79-4093-8cce-667015c1de6d",
		Name:            "Boreas Charger",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Lands: 1, Tutors: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Boreas Charger — choose an opponent who controls more lands than you; search for that many Plains cards", boreasChargerLeaves),
		},
	})
}

func rprB27LandCount(ctx *Context, player uuid.UUID) int {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}

func boreasChargerLeaves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mine := rprB27LandCount(ctx, ctx.Controller())
	var more []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if rprB27LandCount(ctx, opp) > mine {
			more = append(more, opp)
		}
	}
	if len(more) == 0 {
		return nil
	}
	return ChoosePlayer{
		Only:     more,
		Question: "Boreas Charger — choose an opponent who controls more lands than you",
		ThenKey:  boreasChargerChosen,
	}.Apply(ctx)
}

var boreasChargerChosen = ChoosePlayerThen("option-pick/boreas-charger-search", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	diff := rprB27LandCount(ctx, chosen) - rprB27LandCount(ctx, ctx.Controller())
	if diff <= 0 {
		return nil
	}
	return SearchLibrary{
		Player:    ctx.Controller(),
		Source:    ctx.Source(),
		Predicate: func(c game.Card) bool { return c.IsLand() && c.HasSubtypeWord("Plains") },
		Dest:      game.ZoneLibrary,
		ToTop:     true,
		Limit:     diff,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Boreas Charger — a number of Plains cards equal to the difference",
		ThenTail:  game.KeyedTail{Ref: boreasChargerPut, Player: ctx.Controller(), Source: ctx.Source()},
	}.Apply(ctx)
})

// boreasChargerPut is the search's tail: one of the found cards onto the
// battlefield tapped, the rest into hand.
var boreasChargerPut = CardTail("tail/boreas-charger-put", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   ctx.Controller(),
		Cards:    o.IDs,
		Max:      1,
		Tapped:   true,
		Label:    "Boreas Charger — put one of them onto the battlefield tapped (the rest go into your hand)",
		ThenTail: putRestIntoHandTail,
	}.Apply(ctx)
})
