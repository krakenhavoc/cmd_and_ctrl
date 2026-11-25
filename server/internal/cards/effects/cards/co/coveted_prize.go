package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coveted Prize — Sorcery {4}{B} (Zendikar Rising, S82 OLD #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)
//	 Search your library for a card, put it into your hand, then shuffle. If
//	 you have a full party, you may cast a spell with mana value 4 or less
//	 from your hand without paying its mana cost."
//
// The party is counted as the spell is cast (CR 700.8). The search is any
// card to the hand; the free cast is offered once the search has settled,
// if the party is full as it resolves (CR 700.8c), over any spell in the
// caster's hand with mana value 4 or less, the tutored card included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "56204530-7ab5-4610-9570-fba2673725ad",
		Name:              "Coveted Prize",
		Completeness:      effects.CompletenessFull,
		Purpose:           game.Purpose{Tutors: 1},
		SelfCostModifiers: []game.CostModifier{effects.S82OldB089PartyCostReduction()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:   ctx.Controller(),
				Dest:     game.ZoneHand,
				Limit:    1,
				Shuffle:  true,
				Reason:   "Coveted Prize — a card",
				ThenTail: ctx.Tail(covetedPrizeTail),
			}.Apply(ctx)
		},
	})
}

var covetedPrizeFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/coveted-prize-cast", "Coveted Prize")

var covetedPrizeTail = effects.CardTail("then/coveted-prize-free-cast", func(ctx *effects.Context, _ game.TailOutcome) error {
	if effects.FinB9HasFullParty(ctx.Game, ctx.Controller()) {
		effects.OfferFreeHandCast(ctx, ctx.Controller(), covetedPrizeFreeCast,
			"a spell with mana value 4 or less", effects.SpellWithManaValueAtMost(4))
	}
	return nil
})
