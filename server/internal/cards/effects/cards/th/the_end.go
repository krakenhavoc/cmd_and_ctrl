package cardsth

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The End — Instant {2}{B}{B} (Wilds of Eldraine, #4471):
//
//	"This spell costs {2} less to cast if your life total is 5 or less.
//	 Exile target creature or planeswalker. Search its controller's
//	 graveyard, hand, and library for any number of cards with the same
//	 name as that permanent and exile them. That player shuffles, then
//	 draws a card for each card exiled from their hand this way."
//
// The discount is a self cost modifier (CR 601.2f) that holds while the
// caster's life is 5 or less. The body is the extraction family's
// (exilePermanentThenExtract): the name and controller are the permanent's
// as it last existed on the battlefield, read before it is exiled, and
// the search happens whether or not the permanent reached exile. "Any
// number" is Unbounded, so the caster may leave matches in the hidden
// zones. What is new is the tail: the victim's hand is recorded as the
// search starts, the library shuffles (game.runSearchThen runs after the
// shuffle, CR 701.24), and the victim then draws one card for each exiled
// card that was in their hand, as the ability reads it, not the number of
// copies they had.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "835c195a-8c35-4076-896d-8d2972cac270",
		Name:         "The End",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if your life total is 5 or less.", effects.YourLifeIsAtMost(5)),
		},
		Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			victim, names := c.Controller, effects.ExtractionNames(c)
			return effects.ExileTarget{
				Target:   target,
				ThenTail: effects.WithCarry(effects.WithWord(ctx.ItemTail(theEndExiled), strings.Join(names, "\n")), victim),
			}.Apply(ctx)
		},
	})
}

// theEndExiled is the clause after the exile: search the victim's three
// zones for cards sharing a name, exile them, shuffle, then the victim
// draws for each card taken from their hand. The Carry is the victim
// followed by the ids of the cards in their hand as the search begins.
var theEndExiled = effects.CardTail("the-end/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	if victim == uuid.Nil {
		return nil
	}
	names := strings.Split(o.Tail.Word, "\n")
	var hand []uuid.UUID
	if p := ctx.Game.PlayerByIDForEffect(victim); p != nil && p.Hand != nil {
		for _, hc := range p.Hand.Cards {
			hand = append(hand, hc.InstanceID)
		}
	}
	return effects.SearchLibrary{
		Player:               ctx.Controller(),
		LibraryOwner:         victim,
		Predicate:            effects.SharesANameWith(names),
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "The End — search their graveyard, hand and library for cards named " + strings.Join(names, " or "),
		ThenTail:             effects.WithCarry(ctx.ItemTail(theEndDraws), append([]uuid.UUID{victim}, hand...)...),
	}.Apply(ctx)
})

// theEndDraws is the last sentence: the victim draws a card for each found
// card that was in their hand when the search began.
var theEndDraws = effects.CardTail("the-end/draw-for-hand", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	victim, inHand := o.Tail.Carry[0], map[uuid.UUID]bool{}
	for _, id := range o.Tail.Carry[1:] {
		inHand[id] = true
	}
	n := 0
	for _, id := range o.IDs {
		if inHand[id] {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: victim, N: n}.Apply(ctx)
})
