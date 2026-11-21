package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Threats Undetected — Sorcery {2}{G} (Dominaria United):
//
//	"Search your library for up to four creature cards with different
//	 powers and reveal them. An opponent chooses two of those cards.
//	 Shuffle the chosen cards into your library and put the rest into
//	 your hand."
//
// Elemental Teachings' search and its opponent's pick, with the set rule
// "different powers" (searchRuleDifferentPowers) enforced on the picks
// themselves, so a set with two equal powers is refused when the search
// is answered and never offered to a bot seat. The found cards wait in
// the library, revealed, while the controller names the opponent who
// chooses (ChoosePlayer, Emergent Ultimatum's untargeted "an opponent");
// that opponent picks two, clamped to what was found. The rest go to the
// hand and the library is shuffled, which shuffles the chosen cards in
// with it (CR 701.20 shuffle of a library that holds them). A search that
// finds nothing chooses nothing and shuffles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24108160-e72b-4451-b1b1-b2b3ac3ede85",
		Name:         "Threats Undetected",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   ctx.Controller(),
				Source:   ctx.Source(),
				Pred:     func(c game.Card) bool { return c.IsCreature() },
				Dest:     game.ZoneLibrary,
				Limit:    4,
				Reveal:   true,
				Shuffle:  false,
				ToTop:    true,
				Reason:   "Threats Undetected — search your library for up to four creature cards with different powers",
				Rule:     effects.SearchRuleDifferentPowers,
				ThenTail: ctx.ItemTail(threatsUndetectedChooseTail),
			})
		},
	})
}

// threatsUndetectedChooseTail has the caster name the opponent who
// chooses. An on-disk key.
var threatsUndetectedChooseTail = effects.CardTail("threats-undetected/choose", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Threats Undetected — choose an opponent to pick two of the revealed cards to shuffle into your library",
		ThenKey:  threatsUndetectedOpponent,
		Carry:    o.IDs,
	}.Apply(ctx)
})

// threatsUndetectedOpponent is the answer: that opponent picks two.
// Nobody to name chooses no card, so every found card goes to the hand.
// An on-disk key.
var threatsUndetectedOpponent = effects.ChoosePlayerThen("option-pick/threats-undetected-opponent", func(ctx *effects.Context, chosen uuid.UUID, found []uuid.UUID) error {
	if chosen == uuid.Nil {
		return threatsUndetectedSettle(ctx, nil, found)
	}
	return effects.RevealPick{
		Player:   chosen,
		Owner:    ctx.Controller(),
		Question: "Threats Undetected — choose two of these cards; they are shuffled into their library and the rest go to their hand",
		Cards:    found,
		Min:      2,
		Max:      2,
		Keyed:    threatsUndetectedPicked,
	}.Apply(ctx)
})

var threatsUndetectedPicked = effects.RegisterRevealPick("choose-cards/threats-undetected-settle", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	return threatsUndetectedSettle(ctx, picked, left)
})

// threatsUndetectedSettle puts the cards not chosen into the hand, then
// shuffles the library with the chosen ones in it.
func threatsUndetectedSettle(ctx *effects.Context, picked, left []uuid.UUID) error {
	for _, id := range left {
		if ctx.Game.FindCardZoneForEffect(id) == nil {
			continue
		}
		if err := ctx.Game.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	if len(picked) == 0 {
		return nil
	}
	return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
}
