package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herigast, Erupting Nullkite — Legendary Creature — Eldrazi Dragon {9},
// 6/6:
//
//	"Emerge {6}{R}{R} (You may cast this spell by sacrificing a creature
//	 and paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, you may exile your hand. If you do, draw
//	 three cards.
//	 Flying
//	 Each creature spell you cast has emerge. The emerge cost is equal to
//	 its mana cost."
//
// Its own emerge is the shared alternative cost (ADR 0135 §4). The cast
// trigger resolves above Herigast, so it happens even if Herigast is
// countered; the "you may" is asked as it resolves (CR 608.2), and a Yes
// exiles every card in the hand, none included (the ruling: "You may
// choose to exile your hand even if you have no cards in hand. If you do,
// you'll still draw three cards."), then draws three once the exile has
// finished. The last ability is a granted alternative cost (ADR 0118 §3,
// ADR 0135 PR 6): every creature spell its controller casts may be cast
// by sacrificing a creature and paying that spell's own mana cost
// reduced by the creature's mana value, beside any emerge the spell
// prints. Herigast itself may be the creature sacrificed (the ruling:
// losing control of it once the cast has begun doesn't matter).
//
// Caveat: a creature with {X} in its mana cost keeps the {X} in its
// emerge cost, and X is chosen as usual (CR 107.3a), but the engine
// takes a generic cost reduction off the cost's printed generic part
// only and never off the mana announced for X, so the sacrificed
// creature's mana value does not pay for X. CR 107.3a and 601.2f count
// the announced X in the total cost the reduction applies to. That is
// the engine's rule for every cost reduction, not Herigast's alone, and
// it errs toward costing more (#2701).
func init() {
	Register(Spec{
		OracleID:                "76243b38-cab1-465f-aa7d-5bc617541753",
		Name:                    "Herigast, Erupting Nullkite",
		Completeness:            CompletenessCaveats,
		Caveats:                 []string{"When you emerge a creature with {X} in its mana cost, the sacrificed creature's mana value doesn't reduce the mana you pay for X."},
		PrintedKeywords:         []string{"flying"},
		AlternativeCosts:        []game.AlternativeCost{Emerge("{6}{R}{R}")},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{EmergeForCreatureSpellsYouCast()},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Herigast, Erupting Nullkite — you may exile your hand and draw three cards", herigastAsk),
		},
	})
}

// herigastAsk is the cast trigger: the controller decides as it
// resolves.
func herigastAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return MayChoice{
		Question: "Herigast, Erupting Nullkite — exile your hand and draw three cards?",
		YesLabel: "Exile my hand, draw three",
		NoLabel:  "Keep my hand",
		OnYes:    herigastExileHandThenDraw,
	}.Apply(ctx)
}

// herigastExileHandThenDraw exiles the controller's hand as one event
// and draws three when it has finished, however many cards it held.
func herigastExileHandThenDraw(ctx *Context) error {
	controller := ctx.Controller()
	item := ctx.Item
	return ctx.Game.ExileCardsThenForEffect(allHandCardIDs(ctx.Game, controller), func(g *game.Game, _ []uuid.UUID) error {
		return DrawCards{Player: controller, N: 3}.Apply(NewContext(g, item))
	})
}
