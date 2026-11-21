package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, Architect of Thought — Legendary Planeswalker — Jace {2}{U}{U},
// loyalty 4:
//
//	"+1: Until your next turn, whenever a creature an opponent controls
//	 attacks, it gets -1/-0 until end of turn.
//	 −2: Reveal the top three cards of your library. An opponent
//	 separates those cards into two piles. Put one pile into your hand and
//	 the other on the bottom of your library in any order.
//	 −8: For each player, search that player's library for a nonland card
//	 and exile it, then that player shuffles. You may cast those cards
//	 without paying their mana costs."
//
// The +1 is a CR 603.7b delayed trigger with a stated duration (Garruk,
// Curse Breaker's −4 is the template). The engine emits one attack event
// per attacking creature (Actor is its controller), so the trigger goes on
// the stack once per attacker and each resolution shrinks that attacker.
//
// The −2 is Fact or Fiction's reveal, opponent choice and pile pick over
// three cards; the pile not taken goes to the bottom of the library in an
// order the controller chooses (Impulse's ordering prompt).
//
// The −8 searches each player's library in turn order from the controller
// (the controller searches every library, Bribery's LibraryOwner) for a
// nonland card, exiles it and shuffles that library; when every library has
// been searched the controller may cast the exiled cards without paying
// their mana costs (CR 608.2g, Bring to Light's grant). Each search is its
// own prompt, so the progress rides the tail's carry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2274c7ae-5a40-4fd4-a4ac-6f56b23034e4",
		Name:            "Jace, Architect of Thought",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Until your next turn, whenever a creature an opponent controls attacks, it gets -1/-0 until end of turn.",
				Cost:   effects.LoyaltyCost(1),
				Effect: jaceArchitectPlusOne,
			},
			{
				Label: "−2: Reveal the top three cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other on the bottom of your library in any order.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return jaceArchitectReveal(effects.NewContext(g, item))
				},
			},
			{
				Label:  "−8: For each player, search that player's library for a nonland card and exile it, then that player shuffles. You may cast those cards without paying their mana costs.",
				Cost:   effects.LoyaltyCost(-8),
				Effect: jaceArchitectUltimate,
			},
		},
	})
}

// --- +1 ----------------------------------------------------------------------

func jaceArchitectPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	controller := ctx.Controller()
	until := effects.DurationUntilYourNextTurn(ctx, controller)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: ctx.Source(),
		Label:        "Jace, Architect of Thought — the attacking creature gets -1/-0",
		On:           []game.EventKind{game.EventAttack},
		Condition:    jaceArchitectOpponentCreatureAttacksCondition,
		Body:         jaceArchitectShrinkBody,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// jaceArchitectOpponentCreatureAttacksCondition is "a creature an opponent
// controls attacks": the attack event's actor, the attacker's controller,
// is a seated player other than the trigger's controller.
var jaceArchitectOpponentCreatureAttacksCondition = game.DelayedCondition("attack/a-creature-an-opponent-controls-attacks",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == dt.Controller {
			return false
		}
		p := g.PlayerByIDForEffect(ev.Actor)
		return p != nil && !p.Eliminated
	})

var jaceArchitectShrinkBody = game.SimpleDelayedBody("jace-architect-of-thought/attacker-gets-minus-one-zero", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if _, ok := g.PermanentRefForEffect(id); !ok {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: id, Power: -1, Label: "Jace, Architect of Thought — -1/-0"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// --- −2 ----------------------------------------------------------------------

// jaceArchitectReveal reveals the top three cards and names the opponent
// who separates them, chaining the pile pick back to the controller.
func jaceArchitectReveal(ctx *effects.Context) error {
	revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 3,
		"Jace, Architect of Thought — reveal the top three cards of your library")
	if len(revealed) == 0 {
		return nil
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Jace, Architect of Thought — choose an opponent to separate these cards into two piles",
		ThenKey:  jaceArchitectSplit,
		Carry:    revealed,
	}.Apply(ctx)
}

// jaceArchitectSplit hands the revealed cards to the chosen opponent. An
// on-disk key: never renamed, never reused.
var jaceArchitectSplit = effects.ChoosePlayerThen("option-pick/s82-rpr-b25-jace-architect-splitter", func(ctx *effects.Context, splitter uuid.UUID, revealed []uuid.UUID) error {
	if splitter == uuid.Nil {
		// Nobody left to separate them (CR 800.4a): one pile, which the
		// controller takes.
		return jaceArchitectSettle(ctx, revealed, nil)
	}
	return effects.PileSplit{
		Splitter:      splitter,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Jace, Architect of Thought — separate these three cards into two piles",
		PickQuestion:  "Jace, Architect of Thought — take one pile into your hand; the other goes on the bottom of your library",
		Cards:         revealed,
		ThenTail:      ctx.Tail(jaceArchitectTake),
	}.Apply(ctx)
})

// jaceArchitectTake is the continuation both prompts lead to. An on-disk
// key: never renamed, never reused.
var jaceArchitectTake = effects.PileTail("then/s82-rpr-b25-jace-architect-take", func(ctx *effects.Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return jaceArchitectSettle(ctx, taken, left)
})

// jaceArchitectSettle puts the taken pile into the controller's hand and
// the other pile on the bottom of their library in an order they choose. A
// card that is no longer where the split found it is skipped.
func jaceArchitectSettle(ctx *effects.Context, toHand, toBottom []uuid.UUID) error {
	g := ctx.Game
	for _, id := range toHand {
		if g.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	var bottom []uuid.UUID
	for _, id := range toBottom {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			bottom = append(bottom, id)
		}
	}
	if len(bottom) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Cards:     bottom,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Jace, Architect of Thought — put the other pile on the bottom of your library in any order",
	})
}

// --- −8 ----------------------------------------------------------------------

// jaceArchitectUltimate starts the per-player searches: every seated,
// non-eliminated player in turn order from the controller.
func jaceArchitectUltimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return jaceArchitectSearchNext(ctx, jaceArchitectPlayersFrom(g, ctx.Controller()), nil)
}

func jaceArchitectPlayersFrom(g *game.Game, first uuid.UUID) []uuid.UUID {
	start := 0
	for i, p := range g.Seats {
		if p != nil && p.ID == first {
			start = i
		}
	}
	var out []uuid.UUID
	for k := 0; k < len(g.Seats); k++ {
		p := g.Seats[(start+k)%len(g.Seats)]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}

// jaceArchitectCarry packs the players still to search, a uuid.Nil marker
// and the cards exiled so far into a tail's carry.
func jaceArchitectCarry(remaining, exiled []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), remaining...)
	out = append(out, uuid.Nil)
	return append(out, exiled...)
}

// jaceArchitectSearchNext searches the next player's library, or, with
// none left, offers the free casts of everything exiled.
func jaceArchitectSearchNext(ctx *effects.Context, remaining, exiled []uuid.UUID) error {
	if len(remaining) == 0 {
		effects.GrantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Jace, Architect of Thought",
			game.TimingFlash, game.LapseStaysInExile, exiled)
		return nil
	}
	owner := remaining[0]
	return effects.SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: owner,
		Limit:        1,
		Dest:         game.ZoneExile,
		Predicate:    func(c game.Card) bool { return !c.IsLand() },
		Shuffle:      true,
		Source:       ctx.Source(),
		Reason:       "Jace, Architect of Thought — search that player's library for a nonland card and exile it",
		ThenTail:     effects.WithCarry(ctx.ItemTail(jaceArchitectSearched), jaceArchitectCarry(remaining[1:], exiled)...),
	}.Apply(ctx)
}

// jaceArchitectSearched is the continuation of one player's search: carry
// the card it exiled forward and search the next player. An on-disk key:
// never renamed, never reused.
//
// It is assigned in init, not at declaration: the continuation starts the
// next search, which names this tail again.
var jaceArchitectSearched game.TailRef

func init() {
	jaceArchitectSearched = effects.CardTail("then/s82-rpr-b25-jace-architect-searched", jaceArchitectSearchedBody)
}

func jaceArchitectSearchedBody(ctx *effects.Context, o game.TailOutcome) error {
	var remaining, exiled []uuid.UUID
	seenMarker := false
	for _, id := range o.Tail.Carry {
		switch {
		case id == uuid.Nil:
			seenMarker = true
		case seenMarker:
			exiled = append(exiled, id)
		default:
			remaining = append(remaining, id)
		}
	}
	exiled = append(exiled, o.IDs...)
	return jaceArchitectSearchNext(ctx, remaining, exiled)
}
