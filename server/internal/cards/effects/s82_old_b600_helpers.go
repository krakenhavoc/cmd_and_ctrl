package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b600_helpers.go — shared bodies for Betrayers of Kamigawa,
// batch 600 (#5602).

// DrawInsteadLookTopThreeTakeOneBottomRest is the body of "If you would draw
// a card, look at the top three cards of your library instead. Put one of
// those cards into your hand and the rest on the bottom of your library in
// any order" (Tomorrow, Azami's Familiar). It is Underrealm Lich's draw
// replacement with the rest going to the bottom: the look is private, the
// pick is mandatory (one card with one or more to take from), the rest are
// put on the bottom in an order the player names, and only then does the rest
// of the replaced draw instruction run (CR 121.6, CR 614.11). The parked draw
// rides the prompts' Carry as data, so a table waiting on either is a restore
// point. With an empty library nothing is looked at and the draw ends.
func DrawInsteadLookTopThreeTakeOneBottomRest(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	take := TakeFromLibraryToHand{
		Player: drawer,
		Cards:  g.LookAtTopOfLibraryForEffect(drawer, 3),
		Max:    1,
		Label:  tomorrowTakeLabel,
	}
	prompt, answer := take.Ask(g, drawer, source)
	if prompt == nil {
		return tomorrowTake(g, source, drawer, take.Cards, answer, done)
	}
	prompt.Zone = ""
	prompt.LiveZone = game.ZoneLibrary
	prompt.ThenKey = tomorrowPick
	prompt.Carry = append([]uuid.UUID{done.Park(g)}, take.Cards...)
	g.QueueChooseCardsForEffect(*prompt)
	return nil
}

const tomorrowTakeLabel = "Tomorrow, Azami's Familiar — put one of the top three cards into your hand (the rest go on the bottom of your library in any order)"

// tomorrowPick is the pick for the hand. An on-disk key.
var tomorrowPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b600-tomorrow-take", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	var picked []uuid.UUID
	if r.Answered {
		picked = r.Picked
	}
	return tomorrowTake(g, r.Source, r.Chooser, r.Carry[1:], picked, game.DrawDoneByID(r.Carry[0]))
})

// tomorrowTake moves what was picked to hand, then the rest of the look goes
// to the bottom and the parked draw is paid. The tail is registered data (the
// parked draw and the look ride its Carry), so a prompt the move pauses on is
// a restore point.
func tomorrowTake(g *game.Game, source, drawer uuid.UUID, looked, picked []uuid.UUID, done game.DrawDone) error {
	tail := game.KeyedTail{Ref: tomorrowTaken, Player: drawer, Source: source,
		Carry: append([]uuid.UUID{done.Park(g)}, looked...)}
	return FinishTake(g, source, drawer, false, tomorrowTakeLabel, tail, picked)
}

// tomorrowTaken is the rest once the pick has reached the hand: the rest of
// the look goes on the bottom in any order, then the parked draw (Carry[0])
// is paid. An on-disk key.
var tomorrowTaken = game.RegisterTail("s82-old-b600/tomorrow-taken", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	res := TakeResult(g, o.Tail.Source, o.Tail.Player, o.Tail.Carry[1:], o.IDs)
	return TakeThenErr(g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    LibraryOrderLabel(g, res.Source, "put the rest on the bottom of your library in any order"),
		ThenTail:  game.KeyedTail{Ref: tomorrowDone, Player: res.Player, Source: res.Source, Carry: []uuid.UUID{o.Tail.Carry[0]}},
	}))
})

// tomorrowDone pays the rest of the replaced draw once the bottom order has
// been settled. An on-disk key.
var tomorrowDone = game.RegisterTail("s82-old-b600/tomorrow-done", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return TakeThenErr(game.DrawDoneByID(o.Tail.Carry[0]).Run(g))
})

// GrantedAltCostXForSamuraiSpells is the on-disk key of Kentaro, the Smiling
// Cat's granted alternative cost ("You may pay {X} rather than pay the mana
// cost for Samurai spells you cast, where X is that spell's mana value"):
// never renamed, never reused.
const GrantedAltCostXForSamuraiSpells = "granted-pay-x-samurai-spells"
