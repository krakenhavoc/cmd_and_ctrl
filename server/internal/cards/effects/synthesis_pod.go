package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synthesis Pod — Artifact {3}{U/P}:
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 {1}{U/P}, {T}, Exile a spell you control: Target opponent reveals
//	 cards from the top of their library until they reveal a card with
//	 mana value equal to 1 plus the exiled spell's mana value. Exile that
//	 card, then that player shuffles. You may cast that exiled card
//	 without paying its mana cost."
//
// The cost's third part is the exile-a-card component read against the
// stack (ExileASpellYouControl, #5655): the activator names one spell
// they control, a copy included, and it leaves the stack for exile
// without resolving and without being countered. The Phyrexian symbol is
// the cost engine's (Birthing Pod's note).
//
// "The exiled spell's mana value" is the spell's as it last existed on
// the stack (CR 608.2h), so an X in its cost counts as the number chosen
// for it (CR 202.3e) although the card now in exile reads X as 0. The
// engine files that record as the spell leaves the stack, and the effect
// reads it through the paid cost (ctx.Exiled).
//
// The reveal stops on the first card with exactly that mana value plus
// one; a library with no such card is revealed whole and shuffled, and
// nothing is exiled. The cards revealed on the way stay in the library.
// The exile, then the shuffle, then the offer: a free cast during the
// resolution (game.CastExiledFreeKeepingForEffect, CR 608.2g), {0} with
// X as 0, and a card the controller declines or can't cast stays in
// exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55e42539-3fe2-4c84-9978-c071d6fc6aa0",
		Name:         "Synthesis Pod",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U/P}, {T}, Exile a spell you control: Target opponent reveals cards from the top of their library until they reveal a card with mana value equal to 1 plus the exiled spell's mana value. Exile that card, then that player shuffles. You may cast that exiled card without paying its mana cost.",
			Cost:    Plus(ManaCost("{1}{U/P}"), TapCost(), ExileASpellYouControl()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  synthesisPodReveal,
		}},
	})
}

// synthesisPodReveal is the Pod's resolution up to the exile.
func synthesisPodReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	exiled := ctx.Exiled()
	if len(exiled) == 0 {
		return nil
	}
	spell, ok := g.LastKnownSpellManaValueForEffect(exiled[0])
	if !ok {
		return nil
	}
	want := spell + 1
	_, hit := revealUntil(ctx, opp, func(c game.Card) bool {
		mv, ok := c.ParsedManaValue()
		return ok && mv == want
	}, "Synthesis Pod — revealed until a card with mana value "+strconv.Itoa(want))
	if hit == uuid.Nil {
		return g.ShuffleLibraryForEffect(opp)
	}
	return g.ExileCardKeyedForEffect(hit, withCarry(ctx.Tail(synthesisPodExiled), opp))
}

// synthesisPodExiled is the rest once the revealed card's exile settled:
// the opponent shuffles, then the controller may cast the card if it
// reached exile. An on-disk key: never renamed, never reused.
var synthesisPodExiled = CardTail("synthesis-pod/exiled", func(ctx *Context, o game.TailOutcome) error {
	if err := ctx.Game.ShuffleLibraryForEffect(tailCarry0(o)); err != nil {
		return err
	}
	if len(o.IDs) == 0 {
		return nil
	}
	return ctx.Game.CastExiledFreeKeepingForEffect(ctx.Controller(), ctx.Source(), o.IDs[0])
})
