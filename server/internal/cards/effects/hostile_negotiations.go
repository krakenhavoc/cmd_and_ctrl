package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hostile Negotiations — Instant {3}{B} (The Brothers' War, S82):
//
//	"Exile the top three cards of your library in a face-down pile, then
//	 exile the top three cards of your library in another face-down pile.
//	 Look at the cards in each pile, then turn a pile of your choice face up.
//	 An opponent chooses one of those piles. Put that pile into your hand and
//	 the other into your graveyard. You lose 3 life."
//
// Abstract Performance's shape. Each pile is exiled so that nobody may look
// at it (game.ExileTopFaceDownUnseenKeyedForEffect, CR 406.3), one after the
// other on keyed tails, so a short library exiles what it has. You then look
// at both piles (game.LookAtFaceDownExileForEffect: you alone may, and keep
// looking until they leave exile) and choose one to turn face up, which
// reveals it to the table. You choose which opponent chooses (only when there
// is more than one); that opponent picks a pile with the face-up cards shown
// and only the size of the face-down pile. The chosen pile goes into your
// hand and the other into your graveyard (its face-down cards turn face up as
// they leave exile), then you lose 3 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5cbd852-c42f-4592-a0bb-a8bffa2dd2f4",
		Name:         "Hostile Negotiations",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.ExileTopFaceDownUnseenKeyedForEffect(ctx.Controller(), 3, ctx.ItemTail(hostileNegotiationsFirstPile))
		},
	})
}

// hostileNegotiationsFirstPile is the second exile, after the first pile
// landed: the first pile rides the next tail. An on-disk key.
var hostileNegotiationsFirstPile = CardTail("hostile-negotiations/after-first-pile", func(ctx *Context, o game.TailOutcome) error {
	next := ctx.ItemTail(hostileNegotiationsSecondPile)
	next.Carry = o.IDs
	return ctx.Game.ExileTopFaceDownUnseenKeyedForEffect(ctx.Controller(), 3, next)
})

// hostileNegotiationsSecondPile lets the controller look at both piles and
// choose the one to turn face up. An on-disk key.
var hostileNegotiationsSecondPile = CardTail("hostile-negotiations/after-second-pile", func(ctx *Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	first := abstractPerformanceStillExiled(g, o.Tail.Carry)
	second := abstractPerformanceStillExiled(g, o.IDs)
	if len(first) == 0 && len(second) == 0 {
		return hostileNegotiationsLoseLife(ctx)
	}
	g.LookAtFaceDownExileForEffect(you, first)
	g.LookAtFaceDownExileForEffect(you, second)
	return PickOption{
		Player:   you,
		From:     you,
		Question: "Hostile Negotiations — turn a pile of your choice face up",
		Options: []game.ChoiceOption{
			{Label: "The first pile (" + strconv.Itoa(len(first)) + " cards)", Cards: first},
			{Label: "The second pile (" + strconv.Itoa(len(second)) + " cards)", Cards: second},
		},
		ThenKey:  hostileNegotiationsFaceUp,
		Carry:    packIDLists(first, second),
		WithItem: true,
	}.Apply(ctx)
})

// hostileNegotiationsFaceUp turns the chosen pile face up and names the
// opponent who chooses. The carry becomes (face-up pile, face-down pile). An
// on-disk key.
var hostileNegotiationsFaceUp = OptionPickThen("option-pick/hostile-negotiations-face-up", func(ctx *Context, r game.OptionPicked) error {
	g := ctx.Game
	first, second := unpackIDPair(r.Carry)
	up, down := first, second
	if r.Index == 1 {
		up, down = second, first
	}
	g.RevealForEffect(game.RevealSpec{Player: ctx.Controller(), Source: ctx.Source(), Reason: "Hostile Negotiations — the face-up pile", Cards: up})
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Hostile Negotiations — choose an opponent to choose one of the piles",
		ThenKey:  hostileNegotiationsChooser,
		Carry:    packIDLists(up, down),
	}.Apply(ctx)
})

// hostileNegotiationsChooser asks the chosen opponent for a pile. An on-disk
// key.
var hostileNegotiationsChooser = ChoosePlayerThen("option-pick/hostile-negotiations-opponent", func(ctx *Context, opp uuid.UUID, carry []uuid.UUID) error {
	if opp == uuid.Nil {
		return hostileNegotiationsLoseLife(ctx)
	}
	up, down := unpackIDPair(carry)
	return PickOption{
		Player:   opp,
		From:     ctx.Controller(),
		Question: "Hostile Negotiations — choose a pile; it goes into its owner's hand and the other into their graveyard",
		Options: []game.ChoiceOption{
			{Label: "The face-up pile (" + strconv.Itoa(len(up)) + " cards)", Cards: up},
			{Label: "The face-down pile (" + strconv.Itoa(len(down)) + " cards)", Cards: down},
		},
		ThenKey:  hostileNegotiationsPile,
		Carry:    carry,
		WithItem: true,
	}.Apply(ctx)
})

// hostileNegotiationsPile puts the chosen pile into the hand and the other
// into the graveyard; the life loss rides the graveyard move's tail. An
// on-disk key.
var hostileNegotiationsPile = OptionPickThen("option-pick/hostile-negotiations-pile", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil {
		return hostileNegotiationsLoseLife(ctx)
	}
	g, you := ctx.Game, ctx.Controller()
	up, down := unpackIDPair(r.Carry)
	taken, left := up, down
	if r.Index == 1 {
		taken, left = down, up
	}
	for _, id := range abstractPerformanceStillExiled(g, taken) {
		if _, err := g.ReturnFromAnyZoneToHandForEffect(you, ctx.Source(), id); err != nil {
			return err
		}
	}
	left = abstractPerformanceStillExiled(g, left)
	next := ctx.ItemTail(hostileNegotiationsLife)
	if len(left) == 0 {
		return next.Run(g, nil)
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(left, next)
})

// hostileNegotiationsLife is "You lose 3 life", once the piles have moved.
// An on-disk key.
var hostileNegotiationsLife = CardTail("hostile-negotiations/lose-life", func(ctx *Context, _ game.TailOutcome) error {
	return hostileNegotiationsLoseLife(ctx)
})

func hostileNegotiationsLoseLife(ctx *Context) error {
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -3)
}
